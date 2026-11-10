'use client';

import { useState, useRef, useEffect, DragEvent, ChangeEvent } from 'react';
import { 
  Upload, 
  FileVideo, 
  Loader2, 
  AlertCircle, 
  CheckCircle2, 
  Copy, 
  Check, 
  X, 
  Send,
  RotateCcw
} from 'lucide-react';

interface UploadZoneProps {
  onUploadSuccess: (videoId: string) => void;
}

type UploadState = 'idle' | 'selected' | 'uploading' | 'success' | 'error';

const UploadZone = ({ onUploadSuccess }: UploadZoneProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [uploadState, setUploadState] = useState<UploadState>('idle');
  const [progress, setProgress] = useState(0);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [videoId, setVideoId] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [copied, setCopied] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const xhrRef = useRef<XMLHttpRequest | null>(null);

  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const handleFileSelect = (file: File) => {
    if (!file.type.startsWith('video/')) {
      setUploadError('Please select a valid video format (.mp4, .mov, .mkv).');
      setUploadState('error');
      return;
    }

    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }

    const objectUrl = URL.createObjectURL(file);
    setSelectedFile(file);
    setPreviewUrl(objectUrl);
    setUploadState('selected');
    setUploadError(null);
    setVideoId(null);
    setProgress(0);
  };

  const handleStartUpload = async () => {
    if (!selectedFile) return;

    setUploadState('uploading');
    setProgress(0);
    setUploadError(null);

    const formData = new FormData();
    formData.append('video', selectedFile);

    try {
      const xhr = new XMLHttpRequest();
      xhrRef.current = xhr;

      const baseUrl = process.env.NEXT_PUBLIC_API_URL;
      const apiUrl = `${baseUrl}/api/v1/ingest`;

      xhr.open('POST', apiUrl);

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const percent = Math.round((event.loaded / event.total) * 100);
          setProgress(percent);
        }
      };

      xhr.onload = () => {
        if (xhr.status === 200 || xhr.status === 201) {
          try {
            const response = JSON.parse(xhr.responseText);
            console.log(response)
            const generatedId = response.data.videoId;

            if (generatedId) {
              setVideoId(generatedId);
              setUploadState('success');
              onUploadSuccess(generatedId);
            } else {
              setUploadError('Server responded, but video_id was missing.');
              setUploadState('error');
            }
          } catch {
            setUploadError('Invalid JSON response from server.');
            setUploadState('error');
          }
        } else {
          setUploadError(`Upload failed with status code ${xhr.status}`);
          setUploadState('error');
        }
      };

      xhr.onerror = () => {
        setUploadError('Network error during upload. Ensure the Go server is running.');
        setUploadState('error');
      };

      xhr.send(formData);
    } catch (err) {
      console.error(err);
      setUploadError('An unexpected client error occurred.');
      setUploadState('error');
    }
  };

  const handleReset = () => {
    if (xhrRef.current && uploadState === 'uploading') {
      xhrRef.current.abort();
    }
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }
    setSelectedFile(null);
    setPreviewUrl(null);
    setUploadState('idle');
    setProgress(0);
    setUploadError(null);
    setVideoId(null);
    setCopied(false);
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileSelect(e.dataTransfer.files[0]);
    }
  };

  const copyVideoId = () => {
    if (!videoId) return;
    navigator.clipboard.writeText(videoId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="space-y-4 font-jsans">
      {/* 1. IDLE STATE: DROPZONE */}
      {uploadState === 'idle' && (
        <div
          onDragOver={(e) => {
            e.preventDefault();
            setIsDragging(true);
          }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          className={`border-2 border-dashed rounded-2xl p-10 md:p-12 text-center cursor-pointer transition-all duration-300 ${
            isDragging
              ? 'border-orange bg-orange/10 shadow-[0_0_25px_rgba(249,115,22,0.2)]'
              : 'border-zinc-800 hover:border-orange/60 bg-zinc-900/70 hover:shadow-[0_0_25px_rgba(249,115,22,0.15)]'
          }`}
        >
          <input
            ref={fileInputRef}
            type="file"
            accept="video/*"
            className="hidden"
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              if (e.target.files?.[0]) handleFileSelect(e.target.files[0]);
            }}
          />

          <div className="flex flex-col items-center justify-center space-y-4">
            <div className="p-4 rounded-2xl bg-dark border border-zinc-800/80 text-orange group-hover:scale-110 transition-transform">
              {isDragging ? (
                <FileVideo className="w-8 h-8 text-pink animate-bounce" />
              ) : (
                <Upload className="w-8 h-8 text-orange" />
              )}
            </div>

            <div>
              <p className="text-base font-outfit font-bold text-white">
                Drop video file here or <span className="text-pink underline decoration-pink/40 underline-offset-4">browse</span>
              </p>
              <p className="text-xs font-jsans-small text-zinc-400 mt-1">Supports MP4, MOV, or MKV up to 2GB</p>
            </div>

            <div className="inline-flex items-center gap-2 px-5 py-2.5 rounded-full bg-pink hover:opacity-90 text-white text-xs font-outfit font-bold shadow-[0_0_20px_rgba(255,51,181,0.4)] transition-all">
              Select Video File
            </div>
          </div>
        </div>
      )}

      {/* 2. SELECTED / PREVIEW / UPLOADING / SUCCESS / ERROR STATES */}
      {uploadState !== 'idle' && (
        <div className="bg-zinc-900/90 border border-zinc-800 rounded-3xl p-5 md:p-6 space-y-5 shadow-2xl">
          {/* TOP BAR WITH FILE DETAILS & RESET BUTTON */}
          <div className="flex items-center justify-between border-b border-zinc-800 pb-4">
            <div className="flex items-center space-x-3 overflow-hidden">
              <div className="p-2.5 rounded-2xl bg-dark border border-zinc-800 text-orange shrink-0">
                <FileVideo className="w-5 h-5" />
              </div>
              <div className="truncate">
                <p className="text-sm font-outfit font-bold text-white truncate max-w-60 md:max-w-xs">
                  {selectedFile?.name}
                </p>
                <p className="text-xs font-mono text-zinc-400">
                  {selectedFile ? formatFileSize(selectedFile.size) : ''}
                </p>
              </div>
            </div>

            <button
              onClick={handleReset}
              className="p-2 rounded-full bg-dark hover:bg-zinc-800 text-zinc-400 hover:text-white transition-colors"
              title="Change Video File"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* LOCAL VIDEO PREVIEW */}
          {previewUrl && (
            <div className="relative aspect-video bg-dark rounded-2xl overflow-hidden border border-zinc-800">
              <video
                src={previewUrl}
                controls
                className="w-full h-full object-contain"
              />
            </div>
          )}

          {/* STATE: SELECTED (WAITING FOR USER TO PRESS SEND) */}
          {uploadState === 'selected' && (
            <div className="flex flex-col sm:flex-row items-center justify-between gap-3 pt-2">
              <button
                onClick={handleReset}
                className="w-full sm:w-auto px-5 py-2.5 rounded-full bg-dark hover:bg-zinc-800 text-zinc-300 font-jsans text-xs font-semibold border border-zinc-800 transition-colors"
              >
                Choose Different File
              </button>

              <button
                onClick={handleStartUpload}
                className="w-full sm:w-auto px-7 py-2.5 bg-pink hover:opacity-95 text-white font-outfit font-bold text-xs rounded-full shadow-[0_0_20px_rgba(255,51,181,0.45)] transition-all flex items-center justify-center space-x-2 transform hover:scale-105 active:scale-95"
              >
                <Send className="w-4 h-4" />
                <span>Send Video</span>
              </button>
            </div>
          )}

          {/* STATE: UPLOADING (LIVE PROGRESS BAR) */}
          {uploadState === 'uploading' && (
            <div className="space-y-3 pt-1">
              <div className="flex items-center justify-between text-xs">
                <span className="font-outfit font-bold text-white flex items-center gap-2">
                  <Loader2 className="w-4 h-4 text-pink animate-spin" />
                  Uploading Video Stream...
                </span>
                <span className="font-mono font-bold text-orange">{progress}%</span>
              </div>

              {/* ELECTRIC PINK PROGRESS BAR */}
              <div className="w-full h-2.5 bg-dark rounded-full overflow-hidden border border-zinc-800">
                <div
                  className="h-full bg-pink shadow-[0_0_15px_rgba(255,51,181,0.6)] transition-all duration-150"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <p className="text-[11px] font-jsans-small text-zinc-400">
                Transcoding will be dispatched automatically upon completion.
              </p>
            </div>
          )}

          {/* STATE: SUCCESS (SHOW VIDEO ID & ACTIONS) */}
          {uploadState === 'success' && (
            <div className="p-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 space-y-3 animate-fadeIn">
              <div className="flex items-center space-x-2 text-emerald-400">
                <CheckCircle2 className="w-5 h-5 shrink-0" />
                <span className="font-outfit font-bold text-sm">Upload Complete & Queue Dispatched!</span>
              </div>

              {videoId && (
                <div className="space-y-1.5">
                  <label className="text-[11px] font-bold uppercase tracking-wider text-zinc-400">
                    Generated Video ID
                  </label>
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      readOnly
                      value={videoId}
                      className="flex-1 bg-dark border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-orange focus:outline-none"
                    />
                    <button
                      onClick={copyVideoId}
                      className="px-4 py-2 bg-pink text-white text-xs font-bold rounded-xl shadow-[0_0_15px_rgba(255,51,181,0.4)] flex items-center space-x-1 shrink-0"
                    >
                      {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                      <span>{copied ? 'Copied' : 'Copy ID'}</span>
                    </button>
                  </div>
                </div>
              )}

              <div className="pt-2 flex justify-end">
                <button
                  onClick={handleReset}
                  className="px-4 py-2 bg-dark hover:bg-zinc-800 text-zinc-300 font-jsans text-xs font-semibold rounded-xl border border-zinc-800 transition-colors flex items-center space-x-1.5"
                >
                  <RotateCcw className="w-3.5 h-3.5" />
                  <span>Upload Another Video</span>
                </button>
              </div>
            </div>
          )}

          {/* STATE: ERROR */}
          {uploadState === 'error' && uploadError && (
            <div className="p-4 rounded-2xl bg-pink/10 border border-pink/30 space-y-3 animate-fadeIn">
              <div className="flex items-center space-x-2 text-pink">
                <AlertCircle className="w-5 h-5 shrink-0" />
                <span className="font-outfit font-bold text-sm">Upload Failed</span>
              </div>
              <p className="text-xs font-jsans text-zinc-300">{uploadError}</p>

              <div className="pt-1 flex items-center space-x-3">
                <button
                  onClick={handleStartUpload}
                  className="px-4 py-2 bg-pink text-white text-xs font-bold rounded-xl shadow-[0_0_15px_rgba(255,51,181,0.4)] flex items-center space-x-1.5"
                >
                  <RotateCcw className="w-3.5 h-3.5" />
                  <span>Retry Upload</span>
                </button>
                <button
                  onClick={handleReset}
                  className="px-4 py-2 bg-dark hover:bg-zinc-800 text-zinc-300 text-xs font-semibold rounded-xl border border-zinc-800 transition-colors"
                >
                  Choose Different File
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default UploadZone;