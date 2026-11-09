'use client';

import { useState } from 'react';
import HlsPlayer from '@/src/components/HlsPlayer';
import VideoRecorder from '@/src/components/VideoRecorder';
import UploadZone from '@/src/components/UploadZone';
import { 
  Video, 
  Upload, 
  Copy, 
  Check, 
  Play, 
  Clock, 
  Eye, 
  Link2, 
  Sparkles, 
  MoreVertical, 
  Film,
  X,
  Share2
} from 'lucide-react';
import { RecordingItem, SAMPLE_RECORDINGS } from '@/src/data/video';

export default function LandingPage() {
  const [recordings, setRecordings] = useState<RecordingItem[]>(SAMPLE_RECORDINGS);
  const [activeVideoId, setActiveVideoId] = useState<string | null>(null);
  const [isRecorderOpen, setIsRecorderOpen] = useState(false);
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const API_BASE = process.env.NEXT_PUBLIC_API_URL;
  const activeStreamUrl = activeVideoId ? `${API_BASE}/api/v1/videos/${activeVideoId}/master.m3u8` : '';
  const activeShareUrl = activeVideoId ? `${typeof window !== 'undefined' ? window.location.origin : ''}/v/${activeVideoId}` : '';

  const copyLink = (id: string, e?: React.MouseEvent) => {
    if (e) e.stopPropagation();

    const shareUrl = `${window.location.origin}/v/${id}`;
    navigator.clipboard.writeText(shareUrl);

    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRecordingFinished = (newId: string) => {
    setIsRecorderOpen(false);

    const newRecording: RecordingItem = {
      id: newId,
      title: `Instant Stream ${newId.slice(-6)}`,
      duration: '00:00',
      createdAt: 'Just now',
      views: 1,
      thumbnail: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=800&auto=format&fit=crop&q=80',
    };

    setRecordings([newRecording, ...recordings]);
    setActiveVideoId(newId);
  };

  return (
    <div className="min-h-screen bg-dark text-zinc-100 font-jsans selection:bg-pink selection:text-white">
      
      {/* INSTAGRAM-INSPIRED SLEEK HEADER */}
      <header className="sticky top-0 z-40 bg-dark/85 backdrop-blur-md border-b border-zinc-800/80 px-6 md:px-12 h-16 flex items-center justify-between">
        <div className="flex items-center space-x-3 cursor-pointer" onClick={() => setActiveVideoId(null)}>
          <div className="p-0.5 rounded-full bg-linear-to-tr from-orange via-pink to-pink">
            <div className="bg-dark p-1.5 rounded-full">
              <Film className="w-5 h-5 text-orange" />
            </div>
          </div>
          <span className="font-outfit font-extrabold text-xl tracking-tight text-white">
            Prism<span className="text-pink">.</span>
          </span>
        </div>

        {/* HEADER ACTIONS */}
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setIsUploadOpen(true)}
            className="px-4 py-2 bg-zinc-900 hover:bg-zinc-800 text-zinc-300 hover:text-white font-jsans font-semibold text-xs rounded-full border border-zinc-800 transition-colors flex items-center space-x-2"
          >
            <Upload className="w-3.5 h-3.5 text-orange" />
            <span>Upload File</span>
          </button>

          {/* PRIMARY PINK CTA */}
          <button
            onClick={() => setIsRecorderOpen(true)}
            className="px-5 py-2 bg-pink hover:opacity-95 text-white font-outfit font-bold text-xs rounded-full shadow-[0_0_20px_rgba(255,51,181,0.45)] transition-all flex items-center space-x-2 transform hover:scale-105 active:scale-95"
          >
            <Video className="w-4 h-4 stroke-[2.5]" />
            <span>Record Video</span>
          </button>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 md:px-12 py-8 space-y-10">
        
        {/* HERO SECTION */}
        {!activeVideoId && (
          <section className="text-center max-w-2xl mx-auto space-y-4 py-6">
            <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-orange/10 border border-orange/30 text-orange text-xs font-semibold">
              <Sparkles className="w-3.5 h-3.5 text-pink" />
              <span>Real-Time Async Media Platform</span>
            </div>
            
            <h1 className="font-outfit font-extrabold text-4xl md:text-5xl text-white tracking-tight leading-tight">
              Record, Transcode & Share <br />
              <span className="text-transparent bg-clip-text bg-linear-to-r from-orange to-pink">
                Instantly via Link
              </span>
            </h1>

            <p className="font-jsans-small text-zinc-400 text-sm md:text-base max-w-lg mx-auto">
              Start recording or upload your video asset. Stop anytime and get an instant shareable link powered by adaptive HLS streaming.
            </p>

            <div className="flex items-center justify-center space-x-4 pt-2">
              <button
                onClick={() => setIsRecorderOpen(true)}
                className="px-6 py-3 bg-pink hover:opacity-95 text-white font-outfit font-bold text-sm rounded-full shadow-[0_0_25px_rgba(255,51,181,0.5)] transition-all flex items-center space-x-2 transform hover:scale-105"
              >
                <Video className="w-4 h-4" />
                <span>Start Recording</span>
              </button>

              <button
                onClick={() => setIsUploadOpen(true)}
                className="px-6 py-3 bg-zinc-900 hover:bg-zinc-800 text-white font-outfit font-bold text-sm rounded-full border border-zinc-800 transition-all flex items-center space-x-2"
              >
                <Upload className="w-4 h-4 text-orange" />
                <span>Upload Asset</span>
              </button>
            </div>
          </section>
        )}

        {/* ACTIVE PLAYER & INSTANT SHARE BAR */}
        {activeVideoId && (
          <section className="space-y-6 animate-fadeIn">
            <div className="flex items-center justify-between border-b border-zinc-800/80 pb-4">
              <div className="flex items-center space-x-3">
                <span className="w-2.5 h-2.5 rounded-full bg-pink animate-pulse" />
                <h2 className="font-outfit font-bold text-xl text-white">Active Stream Playback</h2>
              </div>
              <button
                onClick={() => setActiveVideoId(null)}
                className="text-xs font-semibold text-zinc-400 hover:text-white px-3 py-1.5 rounded-full bg-zinc-900 border border-zinc-800 flex items-center space-x-1"
              >
                <X className="w-3.5 h-3.5" />
                <span>Close Player</span>
              </button>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
              {/* Player Frame */}
              <div className="lg:col-span-8 space-y-4">
                <HlsPlayer src={activeStreamUrl} autoPlay={true} />
              </div>

              {/* Instant Share Sidebar */}
              <div className="lg:col-span-4 bg-zinc-900/80 border border-zinc-800 rounded-3xl p-6 space-y-5">
                <div className="space-y-1">
                  <div className="flex items-center space-x-2">
                    <Share2 className="w-4 h-4 text-pink" />
                    <h3 className="font-outfit font-bold text-white text-base">Share Video</h3>
                  </div>
                  <p className="text-xs font-jsans-small text-zinc-400">
                    Anyone with this link can watch your video stream instantly.
                  </p>
                </div>

                {/* Instant Link Box */}
                <div className="space-y-2">
                  <label className="text-[11px] font-bold uppercase tracking-wider text-zinc-400">
                    Shareable URL
                  </label>
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      readOnly
                      value={activeShareUrl}
                      className="flex-1 bg-dark border border-zinc-800 rounded-2xl px-3 py-2 text-xs font-mono text-orange focus:outline-none"
                    />
                    <button
                      onClick={() => copyLink(activeVideoId)}
                      className="px-4 py-2 bg-pink text-white text-xs font-bold rounded-2xl shadow-[0_0_15px_rgba(255,51,181,0.4)] flex items-center space-x-1 shrink-0 transition-transform active:scale-95"
                    >
                      {copiedId === activeVideoId ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                      <span>{copiedId === activeVideoId ? 'Copied' : 'Copy'}</span>
                    </button>
                  </div>
                </div>

                <div className="border-t border-zinc-800 pt-4 space-y-2 text-xs font-jsans-small text-zinc-400">
                  <div className="flex justify-between">
                    <span>Access Level</span>
                    <span className="font-semibold text-white">Link Only (Unlisted)</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Stream Engine</span>
                    <span className="font-semibold text-orange">Prism HLS Proxy</span>
                  </div>
                </div>
              </div>
            </div>
          </section>
        )}

        {/* MY RECORDINGS LIBRARY */}
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
                onClick={() => setActiveVideoId(video.id)}
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
                      onClick={(e) => copyLink(video.id, e)}
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

      </main>

      {/* RECORDING STUDIO OVERLAY MODAL */}
      {isRecorderOpen && (
        <VideoRecorder
          onClose={() => setIsRecorderOpen(false)}
          onRecordingComplete={handleRecordingFinished}
        />
      )}

      {/* UPLOAD FILE OVERLAY MODAL */}
      {isUploadOpen && (
        <div className="fixed inset-0 z-50 bg-dark/80 backdrop-blur-md flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-3xl p-6 max-w-lg w-full space-y-4 shadow-[0_0_50px_rgba(249,115,22,0.2)] animate-scaleUp">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <div className="flex items-center space-x-2">
                <Upload className="w-5 h-5 text-pink" />
                <h3 className="font-outfit font-bold text-lg text-white">Upload Media File</h3>
              </div>
              <button 
                onClick={() => setIsUploadOpen(false)}
                className="w-8 h-8 rounded-full bg-zinc-800 text-zinc-400 hover:text-white flex items-center justify-center"
              >
                ✕
              </button>
            </div>

            <UploadZone 
              onUploadSuccess={(id) => {
                setIsUploadOpen(false);
                setActiveVideoId(id);
              }}
            />
          </div>
        </div>
      )}

    </div>
  );
}