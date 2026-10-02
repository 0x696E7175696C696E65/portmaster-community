import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  standalone: false,
  name: 'round',
  pure: true,
})
export class RoundPipe implements PipeTransform {
  transform(value: number, roundBy: number) {
    if (isNaN(value)) {
      return NaN
    }

    return Math.floor(value / roundBy) * roundBy
  }
}
