'use client';

import { useState, useRef, useEffect } from 'react';
import { 
  Video, 
  Square, 
  Mic, 
  MicOff, 
  Camera, 
  CameraOff, 
  Copy, 
  Check, 
  Sparkles, 
  X,
  Loader2,
  AlertCircle,
} from 'lucide-react';

interface VideoRecorderProps {
  onClose: () => void;
  onRecordingComplete: (videoId: string) => void;
}

const VideoRecorder = ({ onClose, onRecordingComplete }: VideoRecorderProps) => {
  const videoPreviewRef = useRef<HTMLVideoElement>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const uploadQueueRef = useRef<Promise<void>>(Promise.resolve());
  const chunkIndexRef = useRef<number>(0);
  const streamFailedRef = useRef<boolean>(false);

  const [isRecording, setIsRecording] = useState(false);
  const [isFinalizing, setIsFinalizing] = useState(false);
  const [recordingTime, setRecordingTime] = useState(0);
  const [chunksUploaded, setChunksUploaded] = useState(0);
  const [micEnabled, setMicEnabled] = useState(true);
  const [cameraEnabled, setCameraEnabled] = useState(true);
  const [recordedVideoId, setRecordedVideoId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const apiBase = (process.env.NEXT_PUBLIC_API_URL!).replace(/\/$/, '');
  const siteBase = (process.env.NEXT_PUBLIC_BASE_URL!).replace(/\/$/, '');

  useEffect(() => {
    let active = true;

    async function setupCamera() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: { width: 1280, height: 720 },
          audio: true,
        });
        if (!active) {
          stream.getTracks().forEach((track) => track.stop());
          return;
        }
        streamRef.current = stream;
        if (videoPreviewRef.current) {
          videoPreviewRef.current.srcObject = stream;
        }
      } catch (err) {
        console.error('Failed to access camera/mic:', err);
        setError('Camera or microphone permission denied. Please allow access and retry.');
      }
    }

    setupCamera();

    return () => {
      active = false;
      if (mediaRecorderRef.current && mediaRecorderRef.current.state !== 'inactive') {
        mediaRecorderRef.current.stop();
      }
      if (streamRef.current) {
        streamRef.current.getTracks().forEach((track) => track.stop());
      }
    };
  }, []);

  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isRecording) {
      interval = setInterval(() => {
        setRecordingTime((prev) => prev + 1);
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [isRecording]);

  const getSupportedMimeType = () => {
    const candidates = [
      'video/webm;codecs=vp9,opus',
      'video/webm;codecs=vp8,opus',
      'video/webm',
    ];
    return candidates.find((type) => MediaRecorder.isTypeSupported(type)) || '';
  };

  const startRecording = () => {
    if (!streamRef.current) return;

    const newVideoId = crypto.randomUUID();
    chunkIndexRef.current = 0;
    uploadQueueRef.current = Promise.resolve();
    streamFailedRef.current = false;

    setError(null);
    setChunksUploaded(0);
    setRecordingTime(0);

    const mimeType = getSupportedMimeType();
    const recorder = mimeType
      ? new MediaRecorder(streamRef.current, { mimeType })
      : new MediaRecorder(streamRef.current);

    recorder.ondataavailable = (e) => {
      if (e.data.size > 0 && !streamFailedRef.current) {
        const seq = chunkIndexRef.current++;
        const chunkBlob = e.data;

        uploadQueueRef.current = uploadQueueRef.current
          .then(async () => {
            if (streamFailedRef.current) return;

            const res = await fetch(`${apiBase}/api/v1/ingest/stream/${newVideoId}?seq=${seq}`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/octet-stream' },
              body: chunkBlob,
            });

            if (!res.ok) {
              streamFailedRef.current = true;
              const msg = await res.text();
              throw new Error(msg || `Chunk ${seq} rejected (${res.status})`);
            }

            setChunksUploaded((prev) => prev + 1);
          })
          .catch((err: Error) => {
            setError(`Live stream interrupted: ${err.message}`);
          });
      }
    };

    recorder.onstop = async () => {
      setIsFinalizing(true);
      try {
        // Wait for the final queued chunk to finish writing to FFmpeg stdin
        await uploadQueueRef.current;

        if (streamFailedRef.current) {
          setIsFinalizing(false);
          return;
        }

        const finishRes = await fetch(`${apiBase}/api/v1/ingest/stream/${newVideoId}/finish`, {
          method: 'POST',
        });

        if (!finishRes.ok) {
          const msg = await finishRes.text();
          throw new Error(msg || `Failed to finalize stream (${finishRes.status})`);
        }

        setRecordedVideoId(newVideoId);

        const shareUrl = `${siteBase}/v/${newVideoId}`;
        navigator.clipboard?.writeText(shareUrl);
        setCopied(true);
        setTimeout(() => setCopied(false), 3000);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to finalize recording.');
      } finally {
        setIsFinalizing(false);
      }
    };

    recorder.start(1000);
    mediaRecorderRef.current = recorder;
    setIsRecording(true);
  };

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
    }
  };

  const toggleMic = () => {
    if (streamRef.current) {
      const audioTrack = streamRef.current.getAudioTracks()[0];

      if (audioTrack) {
        audioTrack.enabled = !audioTrack.enabled;
        setMicEnabled(audioTrack.enabled);
      }
    }
  };

  const toggleCamera = () => {
    if (streamRef.current) {
      const videoTrack = streamRef.current.getVideoTracks()[0];

      if (videoTrack) {
        videoTrack.enabled = !videoTrack.enabled;
        setCameraEnabled(videoTrack.enabled);
      }
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const shareUrl = recordedVideoId ? `${siteBase}/v/${recordedVideoId}` : '';

  return (
    <div className="fixed inset-0 z-50 bg-dark/90 backdrop-blur-md flex items-center justify-center p-4 font-jsans">
      <div className="bg-zinc-900 border border-zinc-800 rounded-3xl overflow-hidden max-w-2xl w-full space-y-0 shadow-[0_0_50px_rgba(249,115,22,0.2)] animate-scaleUp">
        
        {/* HEADER BAR */}
        <div className="p-4 bg-dark border-b border-zinc-800/80 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <span className={`w-2.5 h-2.5 rounded-full ${isRecording ? 'bg-pink animate-ping' : 'bg-orange'}`} />
            <h3 className="font-outfit font-bold text-white text-base">
              {isRecording ? 'Recording Live Stream...' : 'Studio Recording Deck'}
            </h3>
          </div>

          <button
            onClick={onClose}
            disabled={isFinalizing}
            className="w-8 h-8 rounded-full bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white flex items-center justify-center transition-colors disabled:opacity-40"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* CAMERA PREVIEW CANVAS */}
        <div className="relative aspect-video bg-black flex items-center justify-center overflow-hidden">
          <video
            ref={videoPreviewRef}
            autoPlay
            playsInline
            muted
            className="w-full h-full object-cover transform -scale-x-100"
          />

          {/* RECORDING TIMER & LIVE CHUNK COUNTER */}
          {isRecording && (
            <div className="absolute top-4 inset-x-4 flex items-center justify-between">
              <div className="px-3 py-1.5 rounded-full bg-dark/80 backdrop-blur-md border border-pink/40 text-pink text-xs font-mono font-bold flex items-center space-x-2">
                <span className="w-2 h-2 rounded-full bg-pink animate-pulse" />
                <span>REC {formatTime(recordingTime)}</span>
              </div>

              <div className="px-3 py-1.5 rounded-full bg-dark/80 backdrop-blur-md border border-orange/30 text-orange text-[11px] font-mono">
                {chunksUploaded}s streamed
              </div>
            </div>
          )}

          {/* FINALIZING STREAM OVERLAY */}
          {isFinalizing && (
            <div className="absolute inset-0 bg-dark/80 backdrop-blur-sm flex flex-col items-center justify-center space-y-3">
              <Loader2 className="w-9 h-9 text-pink animate-spin" />
              <p className="font-outfit font-bold text-sm text-white">Finalizing HLS Stream...</p>
            </div>
          )}

          {/* AUDIO/VIDEO CONTROLS FLOAT */}
          {!recordedVideoId && !isFinalizing && (
            <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center space-x-3 px-4 py-2 rounded-full bg-dark/80 backdrop-blur-md border border-zinc-800">
              <button
                onClick={toggleMic}
                className={`p-2.5 rounded-full ${micEnabled ? 'bg-zinc-800 text-white' : 'bg-pink/20 text-pink'}`}
              >
                {micEnabled ? <Mic className="w-4 h-4" /> : <MicOff className="w-4 h-4" />}
              </button>

              <button
                onClick={toggleCamera}
                className={`p-2.5 rounded-full ${cameraEnabled ? 'bg-zinc-800 text-white' : 'bg-pink/20 text-pink'}`}
              >
                {cameraEnabled ? <Camera className="w-4 h-4" /> : <CameraOff className="w-4 h-4" />}
              </button>
            </div>
          )}
        </div>

        {/* FOOTER ACTIONS */}
        <div className="p-5 bg-dark space-y-4">
          {error && (
            <div className="p-3 rounded-xl bg-pink/10 border border-pink/30 flex items-center space-x-2 text-pink text-xs">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {!recordedVideoId ? (
            /* RECORD TRIGGER BUTTONS */
            <div className="flex items-center justify-center">
              {!isRecording ? (
                <button
                  onClick={startRecording}
                  disabled={isFinalizing}
                  className="px-8 py-3.5 bg-pink hover:opacity-90 text-white font-outfit font-bold text-base rounded-full shadow-[0_0_25px_rgba(255,51,181,0.5)] transition-all flex items-center space-x-2 transform hover:scale-105 disabled:opacity-50"
                >
                  <Video className="w-5 h-5" />
                  <span>Start Recording</span>
                </button>
              ) : (
                <button
                  onClick={stopRecording}
                  className="px-8 py-3.5 bg-orange hover:opacity-90 text-dark font-outfit font-bold text-base rounded-full shadow-[0_0_25px_rgba(249,115,22,0.4)] transition-all flex items-center space-x-2 transform hover:scale-105"
                >
                  <Square className="w-5 h-5 fill-dark" />
                  <span>Stop & Get Instant Link</span>
                </button>
              )}
            </div>
          ) : (
            /* INSTANT SHARE MODAL STATE */
            <div className="space-y-4 animate-fadeIn">
              <div className="p-3 bg-zinc-900 border border-orange/40 rounded-2xl flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Sparkles className="w-4 h-4 text-pink" />
                  <span className="text-xs font-bold text-white">Video Ready! Link Copied</span>
                </div>
                {copied && <span className="text-xs text-orange font-bold">✓ Copied to clipboard</span>}
              </div>

              {/* LINK & COPY */}
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  readOnly
                  value={shareUrl}
                  className="flex-1 bg-zinc-900 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-orange focus:outline-none"
                />
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(shareUrl);
                    setCopied(true);
                    setTimeout(() => setCopied(false), 2000);
                  }}
                  className="px-4 py-2 bg-pink text-white text-xs font-bold rounded-xl shadow-[0_0_15px_rgba(255,51,181,0.4)] flex items-center space-x-1 shrink-0"
                >
                  {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copied ? 'Copied' : 'Copy'}</span>
                </button>
              </div>

              <button
                onClick={() => onRecordingComplete(recordedVideoId)}
                className="w-full py-2.5 bg-orange hover:opacity-90 text-dark font-outfit font-bold text-sm rounded-xl transition-all"
              >
                Done & Save to Library
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default VideoRecorder;