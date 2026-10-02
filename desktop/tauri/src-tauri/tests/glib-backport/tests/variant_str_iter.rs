use glib::{prelude::*, Variant};

const STRINGS: [&str; 4] = ["first", "", "naïve / 東京", "last"];

fn strings() -> Variant {
    Variant::array_from_iter::<String>(STRINGS.iter().map(|value| value.to_variant()))
}

#[test]
fn next_reads_every_string_including_empty_and_unicode() {
    let variant = strings();
    let mut iter = variant.array_iter_str().unwrap();
    for expected in STRINGS {
        assert_eq!(iter.next(), Some(expected));
    }
    assert_eq!(iter.next(), None);
}

#[test]
fn nth_and_last_read_the_correct_child() {
    let variant = strings();
    let mut iter = variant.array_iter_str().unwrap();
    assert_eq!(iter.nth(2), Some(STRINGS[2]));
    assert_eq!(iter.last(), Some(STRINGS[3]));
    assert_eq!(variant.array_iter_str().unwrap().last(), Some(STRINGS[3]));
    assert_eq!(variant.array_iter_str().unwrap().nth(4), None);
}

#[test]
fn next_back_and_nth_back_read_the_correct_child() {
    let variant = strings();
    let mut iter = variant.array_iter_str().unwrap();
    assert_eq!(iter.next_back(), Some(STRINGS[3]));
    assert_eq!(iter.nth_back(1), Some(STRINGS[1]));
    assert_eq!(iter.next_back(), Some(STRINGS[0]));
    assert_eq!(iter.next_back(), None);
    assert_eq!(variant.array_iter_str().unwrap().nth_back(4), None);
}

#[test]
fn mixed_forward_and_backward_iteration_preserves_bounds() {
    let variant = strings();
    let mut iter = variant.array_iter_str().unwrap();
    assert_eq!(iter.next(), Some(STRINGS[0]));
    assert_eq!(iter.next_back(), Some(STRINGS[3]));
    assert_eq!(iter.len(), 2);
    assert_eq!(iter.next(), Some(STRINGS[1]));
    assert_eq!(iter.next_back(), Some(STRINGS[2]));
    assert_eq!(iter.next(), None);
    assert_eq!(iter.next_back(), None);
}

#[test]
fn empty_string_array_is_exhausted_in_both_directions() {
    let variant = Variant::array_from_iter::<String>([]);
    let mut iter = variant.array_iter_str().unwrap();
    assert_eq!(iter.next(), None);
    assert_eq!(iter.next_back(), None);
    assert_eq!(iter.len(), 0);
}
