export interface RecordingItem {
  id: string;
  title: string;
  duration: string;
  createdAt: string;
  views: number;
  thumbnail: string;
}

export const SAMPLE_RECORDINGS: RecordingItem[] = [
  {
    id: 'rec_8f3a1d9b',
    title: 'Product Overview & Architecture Deep Dive',
    duration: '03:42',
    createdAt: '10 mins ago',
    views: 14,
    thumbnail: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=800&auto=format&fit=crop&q=80',
  },
  {
    id: 'rec_4e9c7a2f',
    title: 'Sprint Demo & Transcoder Performance Benchmark',
    duration: '06:15',
    createdAt: 'Yesterday',
    views: 89,
    thumbnail: 'https://images.unsplash.com/photo-1550745165-9bc0b252726f?w=800&auto=format&fit=crop&q=80',
  },
  {
    id: 'rec_1a2b3c4d',
    title: 'Client Onboarding & Setup Walkthrough',
    duration: '01:50',
    createdAt: '3 days ago',
    views: 204,
    thumbnail: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=800&auto=format&fit=crop&q=80',
  }
];