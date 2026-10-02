import { AppInfo, IntegrationService, PortmasterDir, ProcessInfo } from "./integration";
import { navigationTarget } from './navigation';

export class BrowserIntegrationService implements IntegrationService {
  writeToClipboard(text: string): Promise<void> {
    if (!!navigator.clipboard) {
      return navigator.clipboard.writeText(text);
    }

    return Promise.reject(new Error(`Clipboard API not supported`))
  }

  openExternal(pathOrUrl: string): Promise<void> {
    const destination = navigationTarget(pathOrUrl, location.href);
    if (destination.kind !== 'external') {
      return Promise.reject(new Error('External links must use HTTP or HTTPS'));
    }
    window.open(destination.url, '_blank', 'noopener,noreferrer');
    return Promise.resolve();
  }

  openDir(_: PortmasterDir): Promise<void> {
    return Promise.reject('Not supported in browser')
  }

  getAppIcon(_: ProcessInfo): Promise<string> {
    return Promise.reject('Not supported in browser')
  }

  getAppInfo(_: ProcessInfo): Promise<AppInfo> {
    return Promise.reject('Not supported in browser')
  }

  exitApp(): Promise<void> {
    window.close();

    return Promise.resolve();
  }

  onExitRequest(cb: () => void): () => void {
    // nothing to do, there
    return () => { }
  }
}

