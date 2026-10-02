import { ChangeDetectionStrategy, Component, InjectionToken } from '@angular/core';
import { Overlay } from '@angular/cdk/overlay';
import { MapPin } from './map.service';

// Types retained for map widgets that also serve other application views.
export const MapOverlay = new InjectionToken<Overlay>('MAP_OVERLAY');
export interface Path {
  id: string;
  points: (MapPin | [number, number])[];
  attributes?: { [key: string]: string };
}

/** Community routing setup has no hosted network or account dependencies. */
@Component({
  templateUrl: './spn-page.html',
  styleUrls: ['./spn-page.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SpnPageComponent {}
