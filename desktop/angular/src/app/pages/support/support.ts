import { Component, inject } from '@angular/core';
import { INTEGRATION_SERVICE } from 'src/app/integration';
import { fadeInAnimation, fadeInListAnimation } from 'src/app/shared/animations';
import { SupportType, supportTypes } from './pages';

@Component({
  standalone: false, templateUrl: './support.html', styleUrls: ['./support.scss'], animations: [fadeInAnimation, fadeInListAnimation] })
export class SupportPageComponent {
  readonly supportTypes = supportTypes;
  private readonly integration = inject(INTEGRATION_SERVICE);
  openPage(item: SupportType) {
    if (item.type === 'link') { this.integration.openExternal(item.url); }
  }
}
