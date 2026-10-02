export interface PageSections {
  title?: string;
  choices: SupportType[];
  style?: 'small';
}

export interface QuestionSection {
  title: string;
  help?: string;
}

export interface SupportPage {
  type?: undefined;
  id: string;
  title: string;
  shortHelp: string;
  repoHelp?: string;
  prologue?: string;
  epilogue?: string;
  sections: QuestionSection[];
  privateTicket?: boolean;
  ghIssuePreset?: string;
  includeDebugData?: boolean;
  repositories?: { repo: string, name: string }[];
}

export interface ExternalLink {
  type: 'link',
  url: string;
  title: string;
  shortHelp: string;
}

export type SupportType = SupportPage | ExternalLink;

const repository = 'https://git.frxst.org/bytefrxst/portmaster';
const doc = (file: string) => repository + '/src/branch/main/' + file;
export const supportTypes: PageSections[] = [
  { title: 'Resources', choices: [
    { type: 'link', title: 'Getting started & FAQ', url: doc('docs/community/getting-started.md'), shortHelp: 'Installation, local features, and common questions.' },
    { type: 'link', title: 'Routing & settings guide', url: doc('docs/community/routing.md'), shortHelp: 'Configure applications, Tor, WireGuard, and DNS.' },
    { type: 'link', title: 'Source & release notes', url: repository, shortHelp: 'Browse your community fork and its development history.' },
  ] },
  { title: 'Contribute & get support', style: 'small', choices: [
    { type: 'link', title: 'Issue tracker', url: repository + '/issues', shortHelp: 'Search existing reports and discuss reproducible problems.' },
    { type: 'link', title: 'Contribution guide', url: doc('CONTRIBUTING.md'), shortHelp: 'Build, test, review, and contribute securely.' },
    { type: 'link', title: 'Security & vulnerability reporting', url: doc('SECURITY.md'), shortHelp: 'Read the reporting policy before sharing sensitive details.' },
  ] },
  { title: 'Make a report', style: 'small', choices: [
    { type: 'link', title: 'Report a bug', url: repository + '/issues/new?template=bug_report.md', shortHelp: 'Include your version, expected behavior, and reproduction steps. Review logs before uploading.' },
    { type: 'link', title: 'Suggest an improvement', url: repository + '/issues/new?template=feature_request.md', shortHelp: 'Describe the problem, proposed behavior, and privacy impact.' },
    { type: 'link', title: 'Compatibility report', url: repository + '/issues/new?template=compatibility.md', shortHelp: 'Report operating system, VPN, or application compatibility.' },
  ] },
];
