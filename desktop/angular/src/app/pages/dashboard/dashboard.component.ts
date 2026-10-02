import { KeyValue } from '@angular/common';
import { ChangeDetectionStrategy, ChangeDetectorRef, Component, DestroyRef, OnInit, TrackByFunction, forwardRef, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { BandwidthChartResult, ChartResult, Database, Netquery, SPNService, Verdict } from '@safing/portmaster-api';
import { repeat, retry, timer } from 'rxjs';
import { DefaultBandwidthChartConfig } from 'src/app/shared/netquery/line-chart/line-chart';
import { MAP_HANDLER, MapRef } from '../spn/map-renderer';

interface BlockedProfile {
  profileID: string;
  count: number;
}

interface BandwidthBarData {
  profile: string;
  profile_name: string;

  value: number;
  sent: number;
  received: number;
}


@Component({
  standalone: false,
  selector: 'app-dashboard',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrls: ['./dashboard.component.scss'],
  templateUrl: './dashboard.component.html',
  providers: [
    { provide: MAP_HANDLER, useExisting: forwardRef(() => DashboardPageComponent), multi: true },
  ]
})
export class DashboardPageComponent implements OnInit {

  private readonly destroyRef = inject(DestroyRef);
  private readonly netquery = inject(Netquery);
  private readonly spn = inject(SPNService);
  private readonly cdr = inject(ChangeDetectorRef);


  blockedProfiles: BlockedProfile[] = []

  connectionsPerCountry: {
    [country: string]: number
  } = {};

  get countryNames(): { [country: string]: string } {
    return this.mapRef?.countryNames || {};
  }

  bandwidthLineChart: BandwidthChartResult<any>[] = [];

  bandwidthBarData: BandwidthBarData[] = [];
  statsPending = true;
  statsUnavailable = false;

  bwChartConfig = DefaultBandwidthChartConfig;

  activeConnections: number = 0;
  blockedConnections: number = 0;
  activeProfiles: number = 0;
  dataIncoming = 0;
  dataOutgoing = 0;
  connectionChart: ChartResult[] = [];

  countriesPerProfile: { [profile: string]: string[] } = {}




  features$ = this.spn.watchEnabledFeatures()
    .pipe(takeUntilDestroyed());

  trackCountry: TrackByFunction<KeyValue<string, any>> = (_, ctr) => ctr.key;
  trackApp: TrackByFunction<BlockedProfile> = (_, bp) => bp.profileID;



  private mapRef: MapRef | null = null;

  registerMap(ref: MapRef): void {
    this.mapRef = ref;

    this.mapRef.onMapReady(() => {
      this.updateMapCountries();
    })
  }

  private updateMapCountries() {
    // this check is basically to make typescript happy ...
    if (!this.mapRef) {
      return;
    }

    this.mapRef.worldGroup
      .selectAll('path')
      .classed('active', (d: any) => {
        return !!this.connectionsPerCountry[d.properties.iso_a2];
      });
  }

  unregisterMap(ref: MapRef): void {
    this.mapRef = null;
  }

  onCountryHover(code: string | null) {
    if (!this.mapRef) {
      return
    }

    this.mapRef.worldGroup
      .selectAll('path')
      .classed('hover', (d: any) => {
        return (d.properties.iso_a2 === code);
      });
  }

  onProfileHover(profile: string | null) {
    if (!this.mapRef) {
      return
    }

    this.mapRef.worldGroup
      .selectAll('path')
      .classed('hover', (d: any) => {
        if (!profile) {
          return false;
        }

        return this.countriesPerProfile[profile]?.includes(d.properties.iso_a2);
      });
  }

  ngOnInit() {
    this.netquery
      .batch({
        bwBarChart: {
          query: {
            internal: { $eq: false },
          },
          select: [
            'profile',
            'profile_name',
            {
              $sum: {
                field: 'bytes_sent',
                as: 'sent'
              }
            },
            {
              $sum: {
                field: 'bytes_received',
                as: 'received'
              }
            },
          ],
          groupBy: ['profile', 'profile_name'],
          databases: [Database.Live],
        },

        profileCount: {
          select: [
            'profile',
            {
              $count: {
                field: '*',
                as: 'totalCount'
              }
            }
          ],
          query: {
            verdict: { $in: [Verdict.Block, Verdict.Drop] }
          },
          groupBy: ['profile'],
          databases: [Database.Live]
        },

        countryStats: {
          select: [
            'country',
            { $count: { field: '*', as: 'totalCount' } },
            { $sum: { field: 'bytes_sent', as: 'bwout' } },
            { $sum: { field: 'bytes_received', as: 'bwin' } },
          ],
          query: {
            allowed: { $eq: true },
          },
          groupBy: ['country'],
          databases: [Database.Live]
        },

        perCountryConns: {
          select: ['profile', 'country', 'active', { $count: { field: '*', as: 'totalCount' } }],
          query: {
            allowed: { $eq: true },
          },
          groupBy: ['profile', 'country', 'active'],
          databases: [Database.Live],
        },

      })
      .pipe(
        retry({ delay: () => { this.statsUnavailable = true; this.cdr.markForCheck(); return timer(10000); } }),
        repeat({ delay: 10000 }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(response => {
        this.statsPending = false;
        this.statsUnavailable = false;
        // bandwidth bar chart
        if (response?.bwBarChart){
          const barChartData = response.bwBarChart
            .filter(value => (value.sent + value.received) > 0)
            .sort((a, b) => (b.sent + b.received) - (a.sent + a.received))
            .slice(0, 10);
          this.bandwidthBarData = barChartData.map(row => ({ ...row, profile_name: row.profile_name || 'Unknown application', value: row.sent + row.received })) as BandwidthBarData[]
        }

        // profileCount
        this.blockedConnections = 0;
        this.blockedProfiles = [];

        response.profileCount?.forEach(row => {
          this.blockedConnections += row.totalCount;
          this.blockedProfiles.push({
            profileID: row.profile!,
            count: row.totalCount
          })
        });

        // countryStats
        this.connectionsPerCountry = {};
        this.dataIncoming = 0;
        this.dataOutgoing = 0;

        response.countryStats?.forEach(row => {
          this.dataIncoming += row.bwin;
          this.dataOutgoing += row.bwout;

          if (row.country === '') {
            return
          }

          this.connectionsPerCountry[row.country!] = row.totalCount || 0;
        })

        this.updateMapCountries()

        // perCountryConns
        let profiles = new Set<string>();

        this.activeConnections = 0;
        this.countriesPerProfile = {};

        response.perCountryConns?.forEach(row => {
          if (row.active) {
            profiles.add(row.profile!);
            this.activeConnections += row.totalCount;
          }

          const arr = (this.countriesPerProfile[row.profile!] || []);
          arr.push(row.country!)
          this.countriesPerProfile[row.profile!] = arr;
        });

        this.activeProfiles = profiles.size;

        this.cdr.markForCheck();
      })


    // Charts

    this.netquery
      .activeConnectionChart({})
      .pipe(
        retry({ delay: 10000 }),
        repeat({ delay: 10000 }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(result => {
        this.connectionChart = result;
        this.cdr.markForCheck();
      })

    this.netquery
      .bandwidthChart({}, undefined, 60)
      .pipe(
        retry({ delay: 10000 }),
        repeat({ delay: 10000 }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(bw => {
        this.bandwidthLineChart = bw;
        this.cdr.markForCheck();
      })

  }

}
