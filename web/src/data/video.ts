export interface RecordingItem {
  id: string;
  title: string;
  duration: string;
  createdAt: string;
  thumbnail: string;
}

export const SAMPLE_RECORDINGS: RecordingItem[] = [
  {
    id: 'cdc2db63-6e9e-49c4-9fc6-b56915479318',
    title: 'Hacker',
    duration: '00:05',
    createdAt: '20/11/2026',
    thumbnail: 'https://laws-pray-park-supreme.trycloudflare.com/api/v1/stream/cdc2db63-6e9e-49c4-9fc6-b56915479318/thumbnail.jpg',
  }
];