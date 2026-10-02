// Commands from user space

use num_derive::FromPrimitive;
use num_traits::FromPrimitive;

#[repr(u8)]
#[derive(Clone, Copy, FromPrimitive)]
#[rustfmt::skip]
pub enum CommandType {
    Shutdown              = 0,
    Verdict               = 1,
    UpdateV4              = 2,
    UpdateV6              = 3,
    ClearCache            = 4,
    GetLogs               = 5,
    GetBandwidthStats     = 6,
    PrintMemoryStats      = 7,
    CleanEndedConnections = 8,
}

#[repr(C, packed)]
pub struct Command {
    pub command_type: CommandType,
    value: [u8; 0],
}

#[repr(C, packed)]
#[derive(Debug, PartialEq, Eq)]
pub struct Verdict {
    pub id: u64,
    pub verdict: u8,
}

#[repr(C, packed)]
#[derive(Debug, PartialEq, Eq)]
pub struct UpdateV4 {
    pub protocol: u8,
    pub local_address: [u8; 4],
    pub local_port: u16,
    pub remote_address: [u8; 4],
    pub remote_port: u16,
    pub verdict: u8,
}

#[repr(C, packed)]
#[derive(Debug, PartialEq, Eq)]
pub struct UpdateV6 {
    pub protocol: u8,
    pub local_address: [u8; 16],
    pub local_port: u16,
    pub remote_address: [u8; 16],
    pub remote_port: u16,
    pub verdict: u8,
}

pub fn parse_type(bytes: &[u8]) -> Option<CommandType> {
    FromPrimitive::from_u8(*bytes.first()?)
}

// Decode values explicitly instead of forming references to packed structs
// from a user-controlled buffer. Short payloads never reach a raw pointer read.
pub fn parse_verdict(bytes: &[u8]) -> Option<Verdict> {
    Some(Verdict {
        id: u64::from_le_bytes(bytes.get(..8)?.try_into().ok()?),
        verdict: *bytes.get(8)?,
    })
}

pub fn parse_update_v4(bytes: &[u8]) -> Option<UpdateV4> {
    Some(UpdateV4 {
        protocol: *bytes.first()?,
        local_address: bytes.get(1..5)?.try_into().ok()?,
        local_port: u16::from_le_bytes(bytes.get(5..7)?.try_into().ok()?),
        remote_address: bytes.get(7..11)?.try_into().ok()?,
        remote_port: u16::from_le_bytes(bytes.get(11..13)?.try_into().ok()?),
        verdict: *bytes.get(13)?,
    })
}

pub fn parse_update_v6(bytes: &[u8]) -> Option<UpdateV6> {
    Some(UpdateV6 {
        protocol: *bytes.first()?,
        local_address: bytes.get(1..17)?.try_into().ok()?,
        local_port: u16::from_le_bytes(bytes.get(17..19)?.try_into().ok()?),
        remote_address: bytes.get(19..35)?.try_into().ok()?,
        remote_port: u16::from_le_bytes(bytes.get(35..37)?.try_into().ok()?),
        verdict: *bytes.get(37)?,
    })
}

#[cfg(test)]
use std::fs::File;
#[cfg(test)]
use std::io::Read;
#[cfg(test)]
use std::mem::size_of;
#[cfg(test)]
use std::panic;

#[test]
fn test_go_command_file() {
    let mut file = File::open("testdata/go_command_test.bin").unwrap();
    loop {
        let mut command: [u8; 1] = [0];
        let bytes_count = file.read(&mut command).unwrap();
        if bytes_count == 0 {
            return;
        }
        if let Some(command) = parse_type(&command) {
            match command {
                CommandType::Shutdown => {}
                CommandType::Verdict => {
                    let mut buf = [0; size_of::<Verdict>()];
                    let bytes_count = file.read(&mut buf).unwrap();
                    if bytes_count != size_of::<Verdict>() {
                        panic!("unexpected bytes count")
                    }

                    assert_eq!(parse_verdict(&buf), Some(Verdict { id: 1, verdict: 2 }))
                }
                CommandType::UpdateV4 => {
                    let mut buf = [0; size_of::<UpdateV4>()];
                    let bytes_count = file.read(&mut buf).unwrap();
                    if bytes_count != size_of::<UpdateV4>() {
                        panic!("unexpected bytes count")
                    }

                    assert_eq!(
                        parse_update_v4(&buf),
                        Some(UpdateV4 {
                            protocol: 1,
                            local_address: [1, 2, 3, 4],
                            local_port: 2,
                            remote_address: [2, 3, 4, 5],
                            remote_port: 3,
                            verdict: 4
                        })
                    )
                }
                CommandType::UpdateV6 => {
                    let mut buf = [0; size_of::<UpdateV6>()];
                    let bytes_count = file.read(&mut buf).unwrap();
                    if bytes_count != size_of::<UpdateV6>() {
                        panic!("unexpected bytes count")
                    }

                    assert_eq!(
                        parse_update_v6(&buf),
                        Some(UpdateV6 {
                            protocol: 1,
                            local_address: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16],
                            local_port: 2,
                            remote_address: [
                                2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17
                            ],
                            remote_port: 3,
                            verdict: 4
                        })
                    )
                }
                CommandType::ClearCache => {}
                CommandType::GetLogs => {}
                CommandType::GetBandwidthStats => {}
                CommandType::PrintMemoryStats => {}
                CommandType::CleanEndedConnections => {}
            }
        } else {
            panic!("Unknown command: {}", command[0]);
        }
    }
}

#[test]
fn malformed_commands_are_rejected_without_pointer_reads() {
    assert!(parse_type(&[]).is_none());
    assert!(parse_type(&[255]).is_none());
    for length in 0..core::mem::size_of::<Verdict>() {
        assert!(parse_verdict(&[0u8; 9][..length]).is_none());
    }
    for length in 0..core::mem::size_of::<UpdateV4>() {
        assert!(parse_update_v4(&[0u8; 14][..length]).is_none());
    }
    for length in 0..core::mem::size_of::<UpdateV6>() {
        assert!(parse_update_v6(&[0u8; 38][..length]).is_none());
    }
}

#[test]
fn misaligned_command_payloads_use_little_endian_values() {
    // Each payload starts one byte into its backing allocation, which must
    // not require alignment of the u64/u16 fields in the serialized command.
    let verdict = [0xff, 8, 7, 6, 5, 4, 3, 2, 1, 2];
    assert_eq!(
        parse_verdict(&verdict[1..]),
        Some(Verdict {
            id: 0x0102030405060708,
            verdict: 2
        })
    );
    let ipv4 = [0xff, 6, 127, 0, 0, 1, 0x34, 0x12, 1, 2, 3, 4, 0xcd, 0xab, 4];
    assert_eq!(
        parse_update_v4(&ipv4[1..]),
        Some(UpdateV4 {
            protocol: 6,
            local_address: [127, 0, 0, 1],
            local_port: 0x1234,
            remote_address: [1, 2, 3, 4],
            remote_port: 0xabcd,
            verdict: 4,
        })
    );
    let mut ipv6 = [0u8; 39];
    ipv6[1] = 17;
    ipv6[17] = 1; // Final byte of the local address.
    ipv6[18..20].copy_from_slice(&0x1234u16.to_le_bytes());
    ipv6[35] = 2; // Final byte of the remote address.
    ipv6[36..38].copy_from_slice(&0xabcdu16.to_le_bytes());
    ipv6[38] = 4;
    let update = parse_update_v6(&ipv6[1..]).unwrap();
    let local_port = update.local_port;
    let remote_port = update.remote_port;
    assert_eq!(local_port, 0x1234);
    assert_eq!(remote_port, 0xabcd);
    assert_eq!(update.local_address[15], 1);
    assert_eq!(update.remote_address[15], 2);
    assert_eq!(update.verdict, 4);
}
