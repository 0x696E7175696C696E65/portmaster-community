import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { TauriIntegrationService } from 'src/app/integration/taur-app';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection(),
    {
      provide: TauriIntegrationService,
      useClass: TauriIntegrationService,
      deps: []
    },
  ],
};
