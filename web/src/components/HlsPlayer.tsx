'use client';

import { useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';
import { AlertCircle, Play, Pause, Volume2, VolumeX, Maximize } from 'lucide-react';

interface HlsPlayerProps {
  src: string;
  poster?: string;
  autoPlay?: boolean;
}

const HlsPlayer = ({ src, poster, autoPlay = true }: HlsPlayerProps) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [isMuted, setIsMuted] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !src) return;

    let hls: Hls | null = null;
    setError(null);

    if (Hls.isSupported()) {
      hls = new Hls({
        enableWorker: true,
        lowLatencyMode: true,
        backBufferLength: 90,
      });

      hls.loadSource(src);
      hls.attachMedia(video);

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (autoPlay) {
          video.play().catch(() => {
            // Auto-play was prevented by browser policy (user interaction required)
            setIsPlaying(false);
          });
        }
      });

      hls.on(Hls.Events.ERROR, (_event, data) => {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              setError('Network error encountered while loading video stream.');
              hls?.startLoad();
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              setError('Media playback error. Attempting recovery...');
              hls?.recoverMediaError();
              break;
            default:
              setError('Fatal stream error. Unable to play media.');
              hls?.destroy();
              break;
          }
        }
      });
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = src;
      if (autoPlay) {
        video.play().catch(() => setIsPlaying(false));
      }
    } else {
      setError('HLS video playback is not supported in this browser.');
    }

    return () => {
      if (hls) {
        hls.destroy();
      }
    };
  }, [src, autoPlay]);

  const togglePlay = () => {
    const video = videoRef.current;
    if (!video) return;

    if (video.paused) {
      video.play();
      setIsPlaying(true);
    } else {
      video.pause();
      setIsPlaying(false);
    }
  };

  const toggleMute = () => {
    const video = videoRef.current;
    if (!video) return;

    video.muted = !video.muted;
    setIsMuted(video.muted);
  };

  const toggleFullscreen = () => {
    if (!containerRef.current) return;
    if (!document.fullscreenElement) {
      containerRef.current.requestFullscreen().catch((err) => console.error(err));
    } else {
      document.exitFullscreen().catch((err) => console.error(err));
    }
  };

  return (
    <div
      ref={containerRef}
      className="relative w-full aspect-video rounded-2xl overflow-hidden bg-brand-black border border-zinc-800/80 shadow-2xl group select-none"
    >
      {error && (
        <div className="absolute inset-0 z-20 bg-brand-black/90 backdrop-blur-md flex flex-col items-center justify-center p-6 text-center space-y-3">
          <AlertCircle className="w-10 h-10 text-brand-pink" />
          <p className="text-sm font-semibold text-zinc-200">{error}</p>
        </div>
      )}

      <video
        ref={videoRef}
        poster={poster}
        playsInline
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        className="w-full h-full object-contain cursor-pointer"
        onClick={togglePlay}
      />

      {/* OVERLAY CONTROLS BAR */}
      <div className="absolute bottom-0 inset-x-0 z-10 bg-gradient-to-t from-brand-black/95 via-brand-black/60 to-transparent p-4 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          {/* PLAY / PAUSE BUTTON */}
          <button
            onClick={togglePlay}
            className="w-10 h-10 rounded-xl bg-brand-pink hover:bg-brand-pink-hover text-white flex items-center justify-center shadow-pink-glow transition-all transform active:scale-95"
            aria-label={isPlaying ? 'Pause' : 'Play'}
          >
            {isPlaying ? <Pause className="w-5 h-5 fill-white" /> : <Play className="w-5 h-5 fill-white ml-0.5" />}
          </button>

          {/* VOLUME BUTTON */}
          <button
            onClick={toggleMute}
            className="p-2 text-zinc-300 hover:text-brand-orange transition-colors"
            aria-label={isMuted ? 'Unmute' : 'Mute'}
          >
            {isMuted ? <VolumeX className="w-5 h-5" /> : <Volume2 className="w-5 h-5" />}
          </button>

          <span className="text-xs font-mono font-medium text-brand-orange">Adaptive HLS</span>
        </div>

        {/* FULLSCREEN BUTTON */}
        <button
          onClick={toggleFullscreen}
          className="p-2 text-zinc-300 hover:text-white transition-colors"
          aria-label="Toggle Fullscreen"
        >
          <Maximize className="w-5 h-5" />
        </button>
      </div>
    </div>
  );
}

export default HlsPlayer;