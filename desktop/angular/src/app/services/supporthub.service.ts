
import { Injectable } from '@angular/core';
import { Observable, of, throwError } from 'rxjs';



export interface SupportSection {
  title: string;
  body: string;
}

export interface Issue<CreatedAt = Date> {
  title: string;
  body: string;
  createdAt: CreatedAt;
  repository: string;
  url: string;
  user: string;
  closed?: boolean;
  labels: string[];
}

@Injectable({ providedIn: 'root' })
export class SupportHubService {
  loadIssues(): Observable<Issue[]> { return of([]); }

  uploadText(name: string, content: string): Observable<string> {
    return throwError(() => new Error('Automatic diagnostic uploads are disabled. Review and attach redacted files in the fork issue tracker.'));
  }

  createIssue(repo: string, preset: string, title: string, sections: SupportSection[], debugInfoUrl?: string, opts?: { generateUrl: boolean }): Observable<string> {
    return throwError(() => new Error('Use Get Help to open the fork issue tracker.'));
  }

  createTicket(repoName: string, title: string, email: string, sections: SupportSection[], debugInfoUrl?: string): Observable<void> {
    return throwError(() => new Error('Hosted support tickets are disabled in this community fork.'));
  }
}