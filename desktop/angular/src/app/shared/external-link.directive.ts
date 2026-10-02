import { isPlatformBrowser } from '@angular/common';
import {
  Directive,
  HostBinding, HostListener, Inject,
  Input, OnChanges, PLATFORM_ID, inject
} from '@angular/core';
import { INTEGRATION_SERVICE } from '../integration';
import { navigationTarget } from '../integration/navigation';

@Directive({
  standalone: false,
  // eslint-disable-next-line @angular-eslint/directive-selector
  selector: 'a[href]'
})
export class ExternalLinkDirective implements OnChanges {
  private readonly integration = inject(INTEGRATION_SERVICE);

  @HostBinding('attr.rel')
  relAttr = '';

  @HostBinding('attr.target')
  targetAttr = '';

  @HostBinding('attr.href')
  hrefAttr = '';

  @Input()
  href: string = '';

  constructor(@Inject(PLATFORM_ID) private platformId: string) { }

  @HostListener('click', ['$event'])
  onClick(event: Event) {
    const destination = navigationTarget(this.href, location.href);
    if (destination.kind === 'internal') {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    if (destination.kind === 'external') {
      this.integration.openExternal(destination.url)
        .catch(() => console.error('Unable to open external link'));
    }
  }

  ngOnChanges() {
    const kind = isPlatformBrowser(this.platformId)
      ? navigationTarget(this.href, location.href).kind : 'internal';
    this.hrefAttr = kind === 'blocked' ? '' : this.href;
    this.relAttr = kind === 'external' ? 'noopener noreferrer' : '';
    this.targetAttr = kind === 'external' ? '_blank' : '';
  }
}
