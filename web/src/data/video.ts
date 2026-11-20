export interface RecordingItem {
  id: string;
  title: string;
  duration: string;
  createdAt: string;
  thumbnail: string;
}

export const SAMPLE_RECORDINGS: RecordingItem[] = [
  {
    id: '66e35374-f6be-4ca2-bcf3-629accd36f22',
    title: 'Hacker',
    duration: '00:46',
    createdAt: '20/11/2026',
    thumbnail: 'https://laws-pray-park-supreme.trycloudflare.com/api/v1/stream/66e35374-f6be-4ca2-bcf3-629accd36f22/thumbnail.jpg',
  }
];