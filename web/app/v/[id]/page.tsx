'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import HlsPlayer from '@/src/components/HlsPlayer';
import { 
  Film, 
  Share2, 
  Copy, 
  Check, 
  ArrowLeft, 
  Sparkles, 
  Clock, 
  Eye, 
  ShieldCheck, 
  Video,
  Info
} from 'lucide-react';

export default function WatchPage() {
  const params = useParams();
  const router = useRouter();
  const rawId = params?.id;
  const videoId = Array.isArray(rawId) ? rawId[0] : rawId || '';

  const [copied, setCopied] = useState(false);

  // Construct stream & share endpoints from env or defaults
  const streamBase = (process.env.NEXT_PUBLIC_STREAM_URL!).replace(/\/$/, '');
  const siteBase = (process.env.NEXT_PUBLIC_BASE_URL!).replace(/\/$/, '');

  const streamUrl = videoId ? `${streamBase}/${videoId}/master.m3u8` : '';
  const shareableUrl = videoId ? `${siteBase}/v/${videoId}` : '';

  const copyShareLink = () => {
    if (!shareableUrl) return;
    navigator.clipboard.writeText(shareableUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen bg-dark text-zinc-100 font-jsans selection:bg-pink selection:text-white">
      
      {/* HEADER BAR */}
      <header className="sticky top-0 z-40 bg-dark/85 backdrop-blur-md border-b border-zinc-800/80 px-6 md:px-12 h-16 flex items-center justify-between">
        <Link href="/" className="flex items-center space-x-3 group">
          <div className="p-0.5 rounded-full bg-linear-to-tr from-orange via-pink to-pink group-hover:scale-105 transition-transform">
            <div className="bg-dark p-1.5 rounded-full">
              <Film className="w-5 h-5 text-orange" />
            </div>
          </div>
          <span className="font-outfit font-extrabold text-xl tracking-tight text-white">
            Prism<span className="text-pink">.</span>
          </span>
        </Link>

        {/* PRIMARY CALL TO ACTION */}
        <Link
          href="/"
          className="px-5 py-2 bg-pink hover:opacity-95 text-white font-outfit font-bold text-xs rounded-full shadow-[0_0_20px_rgba(255,51,181,0.45)] transition-all flex items-center space-x-2 transform hover:scale-105 active:scale-95"
        >
          <Video className="w-4 h-4 stroke-[2.5]" />
          <span>Record Your Own</span>
        </Link>
      </header>

      <main className="max-w-6xl mx-auto px-6 md:px-12 py-8 space-y-8">
        
        {/* BREADCRUMB / BACK ACTION */}
        <div className="flex items-center justify-between">
          <button
            onClick={() => router.push('/')}
            className="inline-flex items-center space-x-2 text-xs font-semibold text-zinc-400 hover:text-white transition-colors group"
          >
            <ArrowLeft className="w-4 h-4 text-orange group-hover:-translate-x-1 transition-transform" />
            <span>Back to Studio</span>
          </button>

          <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold">
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>Encrypted HLS Stream</span>
          </div>
        </div>

        {/* MAIN WATCH & SHARE GRID */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          
          {/* PLAYER COLUMN (8 COLS) */}
          <div className="lg:col-span-8 space-y-5">
            <div className="space-y-4">
              <HlsPlayer src={streamUrl} autoPlay={true} />
              
              {/* VIDEO TITLE & METADATA */}
              <div className="p-5 bg-zinc-900/60 border border-zinc-800/80 rounded-3xl space-y-3">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-zinc-800 pb-3">
                  <h1 className="font-outfit font-bold text-xl md:text-2xl text-white tracking-tight">
                    Shared Stream #{videoId ? videoId.slice(0, 8) : '---'}
                  </h1>
                  <span className="text-xs font-mono text-orange px-2.5 py-1 rounded-full bg-dark border border-zinc-800 self-start sm:self-auto">
                    ID: {videoId}
                  </span>
                </div>

                <div className="flex items-center justify-between text-xs font-jsans-small text-zinc-400 pt-1">
                  <div className="flex items-center space-x-4">
                    <span className="flex items-center space-x-1.5">
                      <Clock className="w-3.5 h-3.5 text-orange" />
                      <span>Recorded Stream</span>
                    </span>
                    <span className="flex items-center space-x-1.5">
                      <Eye className="w-3.5 h-3.5 text-pink" />
                      <span>Adaptive 1080p Playback</span>
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* SHARE SIDEBAR (4 COLS) */}
          <div className="lg:col-span-4 bg-zinc-900/90 border border-zinc-800 rounded-3xl p-6 space-y-6 shadow-2xl">
            <div className="space-y-1.5">
              <div className="flex items-center space-x-2">
                <Share2 className="w-5 h-5 text-pink" />
                <h2 className="font-outfit font-bold text-lg text-white">Share Video</h2>
              </div>
              <p className="text-xs font-jsans-small text-zinc-400 leading-relaxed">
                Copy this link to send this recording to anyone. No signup or password required.
              </p>
            </div>

            {/* LINK BOX */}
            <div className="space-y-2">
              <label className="text-[11px] font-bold uppercase tracking-wider text-zinc-400">
                Direct Watch URL
              </label>
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  readOnly
                  value={shareableUrl}
                  className="flex-1 bg-dark border border-zinc-800 rounded-2xl px-3.5 py-2.5 text-xs font-mono text-orange focus:outline-none"
                />
                <button
                  onClick={copyShareLink}
                  className="px-4 py-2.5 bg-pink text-white text-xs font-bold rounded-2xl shadow-[0_0_15px_rgba(255,51,181,0.4)] flex items-center space-x-1 shrink-0 transition-transform active:scale-95"
                >
                  {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copied ? 'Copied!' : 'Copy'}</span>
                </button>
              </div>
            </div>

            {/* TECHNICAL DETAILS */}
            <div className="border-t border-zinc-800 pt-5 space-y-3">
              <div className="flex items-center space-x-2 text-xs font-bold text-white">
                <Info className="w-4 h-4 text-orange" />
                <span>Stream Information</span>
              </div>

              <div className="space-y-2 text-xs font-jsans-small text-zinc-400">
                <div className="flex justify-between items-center py-1 border-b border-zinc-800/50">
                  <span>Privacy Level</span>
                  <span className="font-semibold text-white">Unlisted (Link Access)</span>
                </div>
                <div className="flex justify-between items-center py-1 border-b border-zinc-800/50">
                  <span>Stream Manifest</span>
                  <span className="font-mono text-orange text-[11px]">master.m3u8</span>
                </div>
                <div className="flex justify-between items-center py-1">
                  <span>Proxy Engine</span>
                  <span className="font-semibold text-pink">Prism-go</span>
                </div>
              </div>
            </div>

            {/* RECORD YOUR OWN BANNER */}
            <div className="p-4 rounded-2xl bg-linear-to-br from-orange/10 via-dark to-pink/10 border border-orange/30 space-y-3">
              <div className="flex items-center space-x-2 text-orange font-outfit font-bold text-xs">
                <Sparkles className="w-4 h-4 text-pink" />
                <span>Powered by Prism Engine</span>
              </div>
              <p className="text-xs font-jsans-small text-zinc-300">
                Create instant video recordings with high-throughput HLS transcoding in seconds.
              </p>
              <Link
                href="/"
                className="w-full py-2 bg-orange hover:opacity-95 text-dark font-outfit font-bold text-xs rounded-xl flex items-center justify-center space-x-1 transition-all"
              >
                <span>Create a Video Free</span>
              </Link>
            </div>
          </div>

        </div>

      </main>
    </div>
  );
}