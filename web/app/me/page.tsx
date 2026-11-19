'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import HlsPlayer from '@/src/components/HlsPlayer';
import VideoRecorder from '@/src/components/VideoRecorder';
import UploadZone from '@/src/components/UploadZone';
import OtpAuthForm from '@/app/signin/page';
import {
  Film,
  Video,
  Upload,
  Link2,
  Check,
  Trash2,
  Play,
  Clock,
  LogOut,
  X,
  Loader2,
  FolderOpen,
} from 'lucide-react';

interface UserVideo {
  id: string;
  user_id: string;
  title: string;
  status: string;
  created_at: string;
}

const page = () => {
  const [user, setUser] = useState<{ id: string; email: string } | null>(null);
  const [authChecked, setAuthChecked] = useState(false);

  const [videos, setVideos] = useState<UserVideo[]>([]);
  const [loadingVideos, setLoadingVideos] = useState(false);

  const [activeVideoId, setActiveVideoId] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const [isRecorderOpen, setIsRecorderOpen] = useState(false);
  const [isUploadOpen, setIsUploadOpen] = useState(false);

  const apiBase = (process.env.NEXT_PUBLIC_API_URL!).replace(/\/$/, '');
  const streamBase = (process.env.NEXT_PUBLIC_STREAM_URL!).replace(/\/$/, '');
  const siteBase = (process.env.NEXT_PUBLIC_BASE_URL!).replace(/\/$/, '');

  const fetchMyVideos = useCallback(async () => {
    setLoadingVideos(true);
    try {
      const res = await fetch(`${apiBase}/api/v1/me/videos`, {
        credentials: 'include',
      });

      if (res.ok) {
        const json = await res.json();
        setVideos(Array.isArray(json) ? json : json.data ?? []);
      }
    } finally {
      setLoadingVideos(false);
    }
  }, [apiBase]);

  // Check active JWT cookie session on mount
  useEffect(() => {
    async function checkSession() {
      try {
        const res = await fetch(`${apiBase}/api/v1/auth/me`, {
          credentials: 'include',
        });

        if (res.ok) {
          const json = await res.json();
          setUser(json.data ?? json);
          await fetchMyVideos();
        }
      } finally {
        setAuthChecked(true);
      }
    }
    checkSession();
  }, [apiBase, fetchMyVideos]);

  const handleLogout = async () => {
    await fetch(`${apiBase}/api/v1/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });

    setUser(null);
    setVideos([]);
    setActiveVideoId(null);
  };

  const copyShareLink = (id: string, e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    navigator.clipboard.writeText(`${siteBase}/v/${id}`);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirmDeleteId !== id) {
      setConfirmDeleteId(id);
      setTimeout(() => setConfirmDeleteId((prev) => (prev === id ? null : prev)), 3000);
      return;
    }

    setDeletingId(id);

    try {
      const res = await fetch(`${apiBase}/api/v1/videos/${id}`, {
        method: 'DELETE',
        credentials: 'include',
      });
      if (res.ok) {
        setVideos((prev) => prev.filter((v) => v.id !== id));
        if (activeVideoId === id) setActiveVideoId(null);
      }
    } finally {
      setDeletingId(null);
      setConfirmDeleteId(null);
    }
  };

  const handleCreatedVideo = async (newId: string) => {
    setIsRecorderOpen(false);
    setIsUploadOpen(false);
    setActiveVideoId(newId);
    await fetchMyVideos();
  };

  if (!authChecked) {
    return (
      <div className="min-h-screen bg-dark flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-pink animate-spin" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-dark text-zinc-100 font-jsans selection:bg-pink selection:text-white">
      {/* HEADER */}
      <header className="sticky top-0 z-40 bg-dark/85 backdrop-blur-md border-b border-zinc-800/80 px-6 md:px-12 h-16 flex items-center justify-between">
        <Link href="/" className="flex items-center space-x-3">
          <div className="p-0.5 rounded-full bg-linear-to-tr from-orange via-pink to-pink">
            <div className="bg-dark p-1.5 rounded-full">
              <Film className="w-5 h-5 text-orange" />
            </div>
          </div>
          <span className="font-outfit font-extrabold text-xl tracking-tight text-white">
            Prism<span className="text-pink">.</span>
          </span>
        </Link>

        {user && (
          <div className="flex items-center space-x-3">
            <span className="hidden sm:inline-block text-xs font-mono text-zinc-400 bg-zinc-900 border border-zinc-800 px-3 py-1.5 rounded-full">
              {user.email}
            </span>

            <button
              onClick={() => setIsUploadOpen(true)}
              className="px-3.5 py-2 bg-zinc-900 hover:bg-zinc-800 text-zinc-300 hover:text-white text-xs font-semibold rounded-full border border-zinc-800 flex items-center space-x-1.5"
            >
              <Upload className="w-3.5 h-3.5 text-orange" />
              <span>Upload</span>
            </button>

            <button
              onClick={() => setIsRecorderOpen(true)}
              className="px-4 py-2 bg-pink hover:opacity-95 text-white font-outfit font-bold text-xs rounded-full shadow-[0_0_20px_rgba(255,51,181,0.45)] flex items-center space-x-1.5"
            >
              <Video className="w-3.5 h-3.5" />
              <span>Create New</span>
            </button>

            <button
              onClick={handleLogout}
              title="Sign Out"
              className="p-2 rounded-full bg-zinc-900 hover:bg-zinc-800 text-zinc-400 hover:text-white border border-zinc-800"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        )}
      </header>

      <main className="max-w-6xl mx-auto px-6 md:px-12 py-10 space-y-10">
        {!user ? (
          <div className="py-12">
            <OtpAuthForm
              onAuthenticated={async (authedUser) => {
                setUser(authedUser);
                await fetchMyVideos();
              }}
            />
          </div>
        ) : (
          <>
            {/* INLINE PLAYER WHEN A VIDEO IS SELECTED */}
            {activeVideoId && (
              <section className="space-y-4">
                <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
                  <h2 className="font-outfit font-bold text-lg text-white">Now Playing</h2>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => copyShareLink(activeVideoId)}
                      className="px-3.5 py-1.5 rounded-full bg-pink text-white text-xs font-bold flex items-center space-x-1.5"
                    >
                      {copiedId === activeVideoId ? <Check className="w-3.5 h-3.5" /> : <Link2 className="w-3.5 h-3.5" />}
                      <span>{copiedId === activeVideoId ? 'Link Copied' : 'Copy Share Link'}</span>
                    </button>
                    <button
                      onClick={() => setActiveVideoId(null)}
                      className="p-1.5 rounded-full bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                </div>
                <HlsPlayer src={`${streamBase}/${activeVideoId}/master.m3u8`} autoPlay={true} />
              </section>
            )}

            {/* USER VIDEO LIBRARY HEADER */}
            <section className="space-y-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <h1 className="font-outfit font-extrabold text-2xl text-white">My Videos</h1>
                  <span className="text-xs font-mono px-2.5 py-0.5 rounded-full bg-zinc-900 border border-zinc-800 text-orange">
                    {videos.length}
                  </span>
                </div>
              </div>

              {loadingVideos ? (
                <div className="py-16 flex justify-center">
                  <Loader2 className="w-7 h-7 text-orange animate-spin" />
                </div>
              ) : videos.length === 0 ? (
                /* EMPTY STATE */
                <div className="border-2 border-dashed border-zinc-800 rounded-3xl p-12 text-center space-y-4 bg-zinc-900/40">
                  <FolderOpen className="w-10 h-10 text-orange mx-auto" />
                  <div className="space-y-1">
                    <p className="font-outfit font-bold text-base text-white">No recordings yet</p>
                    <p className="text-xs font-jsans-small text-zinc-400">
                      Record your screen/camera or upload a file to generate an instant share link.
                    </p>
                  </div>
                  <div className="flex items-center justify-center space-x-3 pt-2">
                    <button
                      onClick={() => setIsRecorderOpen(true)}
                      className="px-5 py-2.5 bg-pink text-white font-outfit font-bold text-xs rounded-full shadow-[0_0_20px_rgba(255,51,181,0.45)] flex items-center space-x-2"
                    >
                      <Video className="w-4 h-4" />
                      <span>Record First Video</span>
                    </button>
                    <button
                      onClick={() => setIsUploadOpen(true)}
                      className="px-5 py-2.5 bg-zinc-900 text-white font-outfit font-bold text-xs rounded-full border border-zinc-800 flex items-center space-x-2"
                    >
                      <Upload className="w-4 h-4 text-orange" />
                      <span>Upload File</span>
                    </button>
                  </div>
                </div>
              ) : (
                /* VIDEO CARDS GRID */
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {videos.length > 0 && videos.map((video) => (
                    <div
                      key={video.id}
                      onClick={() => setActiveVideoId(video.id)}
                      className="group bg-zinc-900/60 border border-zinc-800/80 hover:border-orange/50 rounded-3xl overflow-hidden cursor-pointer transition-all hover:shadow-[0_0_25px_rgba(249,115,22,0.15)] flex flex-col justify-between"
                    >
                      {/* PREVIEW BANNER */}
                      <div className="relative aspect-video bg-linear-to-br from-zinc-900 via-dark to-zinc-950 flex items-center justify-center border-b border-zinc-800/60 overflow-hidden">
                        {video.status === 'READY' && (
                          <img
                            src={`${streamBase}/${video.id}/thumbnail.jpg`}
                            alt={video.title}
                            loading="lazy"
                            onError={(e) => {
                              e.currentTarget.style.display = 'none';
                            }}
                            className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                          />
                        )}

                        <div className="relative z-10 w-12 h-12 rounded-full bg-pink/90 text-white flex items-center justify-center shadow-[0_0_20px_rgba(255,51,181,0.6)] group-hover:scale-110 transition-transform">
                          <Play className="w-5 h-5 fill-white ml-0.5" />
                        </div>

                        <span className="absolute top-3 left-3 z-10 px-2.5 py-0.5 rounded-full bg-dark/80 backdrop-blur-xs border border-zinc-800 text-[10px] font-mono text-orange">
                          {video.status}
                        </span>
                      </div>

                      {/* CARD BODY & ACTIONS */}
                      <div className="p-4 space-y-4">
                        <div>
                          <h3 className="font-outfit font-bold text-sm text-white group-hover:text-orange transition-colors truncate">
                            {video.title}
                          </h3>
                          <p className="text-[11px] font-jsans-small text-zinc-400 flex items-center space-x-1 mt-1">
                            <Clock className="w-3 h-3 text-zinc-500" />
                            <span>{new Date(video.created_at).toLocaleDateString()}</span>
                          </p>
                        </div>

                        <div className="flex items-center justify-between gap-2 pt-1 border-t border-zinc-800/70">
                          {/* SHARE BUTTON */}
                          <button
                            onClick={(e) => copyShareLink(video.id, e)}
                            className="flex-1 py-2 px-3 rounded-xl bg-zinc-800 hover:bg-pink text-zinc-200 hover:text-white text-xs font-bold transition-colors flex items-center justify-center space-x-1.5"
                          >
                            {copiedId === video.id ? <Check className="w-3.5 h-3.5" /> : <Link2 className="w-3.5 h-3.5" />}
                            <span>{copiedId === video.id ? 'Copied Link' : 'Share'}</span>
                          </button>

                          {/* DELETE BUTTON (CLICK TWICE TO CONFIRM) */}
                          <button
                            onClick={(e) => handleDelete(video.id, e)}
                            disabled={deletingId === video.id}
                            className={`py-2 px-3 rounded-xl text-xs font-bold transition-colors flex items-center space-x-1 ${
                              confirmDeleteId === video.id
                                ? 'bg-pink text-white'
                                : 'bg-zinc-900 hover:bg-pink/20 text-zinc-400 hover:text-pink border border-zinc-800'
                            }`}
                            title="Delete video"
                          >
                            {deletingId === video.id ? (
                              <Loader2 className="w-3.5 h-3.5 animate-spin" />
                            ) : (
                              <Trash2 className="w-3.5 h-3.5" />
                            )}
                            {confirmDeleteId === video.id && <span>Confirm</span>}
                          </button>
                        </div>
                      </div>
                    </div>
                    ))}
                  </div>
              )}
            </section>
          </>
        )}
      </main>

      {/* CREATE ANOTHER: RECORDER MODAL */}
      {isRecorderOpen && (
        <VideoRecorder
          onClose={() => setIsRecorderOpen(false)}
          onRecordingComplete={handleCreatedVideo}
        />
      )}

      {/* CREATE ANOTHER: UPLOAD MODAL */}
      {isUploadOpen && (
        <div className="fixed inset-0 z-50 bg-dark/80 backdrop-blur-md flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-3xl p-6 max-w-lg w-full space-y-4 shadow-[0_0_50px_rgba(249,115,22,0.2)]">
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

            <UploadZone onUploadSuccess={handleCreatedVideo} />
          </div>
        </div>
      )}
    </div>
  );
}

export default page;