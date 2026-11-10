'use client';

import { Play, Clock, Eye, Link2, MoreVertical, Check } from 'lucide-react';
import { RecordingItem } from '@/src/data/video';

interface PublicLibraryProps {
  recordings: RecordingItem[];
  onSelectVideo: (id: string) => void;
  copiedId: string | null;
  onCopyLink: (id: string, e?: React.MouseEvent) => void;
}

export default function PublicLibrary({
  recordings,
  onSelectVideo,
  copiedId,
  onCopyLink,
}: PublicLibraryProps) {
  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="font-outfit font-bold text-xl text-white flex items-center space-x-2">
          <span>Public Library</span>
          <span className="text-xs font-mono px-2 py-0.5 rounded-full bg-zinc-900 border border-zinc-800 text-orange">
            {recordings.length}
          </span>
        </h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {recordings.map((video) => (
          <div
            key={video.id}
            onClick={() => onSelectVideo(video.id)}
            className="group bg-zinc-900/60 border border-zinc-800/80 hover:border-orange/50 rounded-3xl overflow-hidden cursor-pointer transition-all duration-300 hover:shadow-[0_0_25px_rgba(249,115,22,0.15)] flex flex-col justify-between"
          >
            {/* Media Thumbnail Frame */}
            <div className="relative aspect-video bg-black overflow-hidden">
              <img
                src={video.thumbnail}
                alt={video.title}
                className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500 opacity-90"
              />

              {/* Play Overlay */}
              <div className="absolute inset-0 bg-dark/40 group-hover:bg-dark/20 transition-colors flex items-center justify-center">
                <div className="w-12 h-12 rounded-full bg-pink/90 text-white flex items-center justify-center shadow-[0_0_20px_rgba(255,51,181,0.6)] transform group-hover:scale-110 transition-transform">
                  <Play className="w-5 h-5 fill-white ml-0.5" />
                </div>
              </div>

              {/* Duration Pill */}
              <span className="absolute bottom-3 right-3 px-2.5 py-1 bg-dark/80 backdrop-blur-md text-[11px] font-mono font-medium text-orange rounded-full border border-zinc-700">
                {video.duration}
              </span>
            </div>

            {/* Meta Information & Quick Copy Button */}
            <div className="p-4 space-y-3">
              <div className="flex items-start justify-between gap-2">
                <h3 className="font-outfit font-bold text-sm text-zinc-100 group-hover:text-orange transition-colors line-clamp-1">
                  {video.title}
                </h3>
                <button 
                  onClick={(e) => {
                    e.stopPropagation();
                  }}
                  className="text-zinc-500 hover:text-zinc-300 p-1"
                >
                  <MoreVertical className="w-4 h-4" />
                </button>
              </div>

              <div className="flex items-center justify-between text-xs text-zinc-400 font-jsans-small">
                <div className="flex items-center space-x-3">
                  <span className="flex items-center space-x-1">
                    <Clock className="w-3.5 h-3.5 text-zinc-500" />
                    <span>{video.createdAt}</span>
                  </span>
                  <span className="flex items-center space-x-1">
                    <Eye className="w-3.5 h-3.5 text-zinc-500" />
                    <span>{video.views}</span>
                  </span>
                </div>

                {/* Instant Link Button */}
                <button
                  onClick={(e) => onCopyLink(video.id, e)}
                  className="px-2.5 py-1 bg-zinc-800 hover:bg-pink hover:text-white text-zinc-300 text-[11px] font-bold rounded-xl transition-all flex items-center space-x-1"
                >
                  {copiedId === video.id ? <Check className="w-3 h-3" /> : <Link2 className="w-3 h-3" />}
                  <span>{copiedId === video.id ? 'Copied' : 'Link'}</span>
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}