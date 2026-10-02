import { AfterViewInit, Directive, ElementRef, HostBinding, Input, OnChanges, Renderer2, SimpleChanges } from '@angular/core';

@Directive({
  standalone: false,
  selector: 'span[appCountryFlags]',
})
export class CountryFlagDirective implements AfterViewInit, OnChanges {
  private readonly flagDir = "/assets/img/flags/";

  @HostBinding('style.text-shadow')
  textShadow = 'rgba(255, 255, 255, .5) 0px 0px 1px';

  @Input()
  appCountryFlags: string = '';

  constructor(
    private el: ElementRef,
    private renderer: Renderer2
  ) { }

  ngOnChanges(changes: SimpleChanges): void {
    if (!changes['appCountryFlags'].isFirstChange()) {
      this.update();
    }
  }

  ngAfterViewInit() {
    this.update();
  }

  private update() {
    const span = this.el.nativeElement as HTMLSpanElement;
    const code = typeof this.appCountryFlags === 'string' ? this.appCountryFlags.toUpperCase() : '';
    const validCode = /^(?:[A-Z]{2}|__)$/.test(code);
    const flag = validCode && code !== '__' ? this.toUnicodeFlag(code) : '🌐';
    this.renderer.setAttribute(span, 'data-before', flag);
    span.replaceChildren();
    if (validCode) {
      const image = span.ownerDocument.createElement('img');
      image.style.display = 'inline';
      image.src = `${this.flagDir}${code}.png`;
      image.alt = code === '__' ? 'Anycast' : code;
      span.appendChild(image);
    }
  }

  private toUnicodeFlag(code: string) {
    const base = 127462 - 65;
    const cc = code.toUpperCase();
    const res = String.fromCodePoint(...cc.split('').map(c => base + c.charCodeAt(0)));
    return res;
  }
}
