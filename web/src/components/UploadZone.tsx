'use client';

import { useState, useRef, DragEvent, ChangeEvent } from 'react';
import { Upload, FileVideo, Loader2, AlertCircle } from 'lucide-react';

interface UploadZoneProps {
  onUploadSuccess: (videoId: string) => void;
}

const UploadZone = ({ onUploadSuccess }: UploadZoneProps) => {
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleUpload = async (file: File) => {
    if (!file.type.startsWith('video/')) {
      setUploadError('Please select a valid video format (.mp4, .mov, .mkv).');
      return;
    }

    setUploadError(null);
    setIsUploading(true);
    setProgress(0);

    const formData = new FormData();
    formData.append('video', file);

    try {
      const xhr = new XMLHttpRequest();
      const apiUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/videos/upload`;
      
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
            onUploadSuccess(response.video_id);
          } catch {
            setUploadError('Invalid JSON response from server.');
          }
        } else {
          setUploadError(`Upload failed with status code ${xhr.status}`);
        }
        setIsUploading(false);
      };

      xhr.onerror = () => {
        setUploadError('Network error during upload. Ensure the Go server is running.');
        setIsUploading(false);
      };

      xhr.send(formData);
    } catch (err) {
      console.error(err);
      setUploadError('An unexpected client error occurred.');
      setIsUploading(false);
    }
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleUpload(e.dataTransfer.files[0]);
    }
  };

  return (
    <div className="space-y-4">
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        onClick={() => !isUploading && fileInputRef.current?.click()}
        className={`border-2 border-dashed rounded-2xl p-10 md:p-14 text-center cursor-pointer transition-all duration-300 ${
          isDragging
            ? 'border-brand-orange bg-brand-orange/10 shadow-orange-glow'
            : 'border-zinc-800 hover:border-brand-orange/60 bg-brand-card/70 hover:shadow-orange-glow'
        }`}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept="video/*"
          className="hidden"
          disabled={isUploading}
          onChange={(e: ChangeEvent<HTMLInputElement>) => {
            if (e.target.files?.[0]) handleUpload(e.target.files[0]);
          }}
        />

        <div className="flex flex-col items-center justify-center space-y-4">
          {isUploading ? (
            <>
              <div className="relative">
                <Loader2 className="w-12 h-12 text-brand-pink animate-spin" />
                <span className="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-white">
                  {progress}%
                </span>
              </div>
              
              <div className="space-y-1">
                <p className="text-sm font-semibold text-zinc-100">Uploading Video File...</p>
                <p className="text-xs text-zinc-400">Transcoding queued automatically upon completion</p>
              </div>

              {/* ELECTRIC PINK PROGRESS BAR */}
              <div className="w-full max-w-xs h-2 bg-zinc-950 rounded-full overflow-hidden border border-zinc-800">
                <div
                  className="h-full bg-brand-pink shadow-pink-glow transition-all duration-150"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </>
          ) : (
            <>
              <div className="p-4 rounded-2xl bg-zinc-950 border border-zinc-800/80 text-brand-orange group-hover:scale-110 transition-transform">
                {isDragging ? (
                  <FileVideo className="w-8 h-8 text-brand-pink animate-bounce" />
                ) : (
                  <Upload className="w-8 h-8 text-brand-orange" />
                )}
              </div>
              
              <div>
                <p className="text-base font-bold text-white">
                  Drop video file here or <span className="text-brand-pink underline decoration-brand-pink/40 underline-offset-4">browse</span>
                </p>
                <p className="text-xs text-zinc-400 mt-1">Supports MP4, MOV, or MKV up to 2GB</p>
              </div>

              <div className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-brand-pink hover:bg-brand-pink-hover text-white text-xs font-bold shadow-pink-glow transition-all">
                Select Video File
              </div>
            </>
          )}
        </div>
      </div>

      {uploadError && (
        <div className="p-4 rounded-xl bg-brand-pink/10 border border-brand-pink/30 flex items-center gap-3 text-brand-pink text-xs font-medium">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{uploadError}</span>
        </div>
      )}
    </div>
  );
}

export default UploadZone;