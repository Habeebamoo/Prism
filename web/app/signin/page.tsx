'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Mail, KeyRound, Loader2, ArrowRight, RotateCcw } from 'lucide-react';

interface OtpAuthFormProps {
  onAuthenticated?: (user: { id: string; email: string }) => void;
}

const Page = ({ onAuthenticated }: OtpAuthFormProps) => {
  const router = useRouter();

  const [step, setStep] = useState<'email' | 'otp'>('email');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const apiBase = (process.env.NEXT_PUBLIC_API_URL!).replace(/\/$/, '');

  const handleRequestOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    let data;

    try {
      const res = await fetch(`${apiBase}/api/v1/auth/request-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      if (!res.ok) {
        data = await res.json()
        throw new Error(data.message || 'Failed to send OTP');
      }

      setStep('otp');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    let data;

    try {
      const res = await fetch(`${apiBase}/api/v1/auth/verify-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, code }),
      });

      if (!res.ok) {
        data = await res.json()
        throw new Error(data.message || 'Invalid verification code');
      }

      data = await res.json();
      onAuthenticated?.(data.data ?? data);
      router.push('/me');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-dark flex items-center justify-center p-4 font-jsans">
      <div className="w-full max-w-md bg-zinc-900/90 border border-zinc-800 rounded-3xl p-7 space-y-6 shadow-[0_0_40px_rgba(249,115,22,0.12)]">
        <div className="space-y-1.5 text-center">
          <h2 className="font-outfit font-extrabold text-2xl text-white">
            {step === 'email' ? 'Access Your Studio' : 'Check Your Inbox'}
          </h2>
          <p className="text-xs font-jsans-small text-zinc-400">
            {step === 'email'
              ? 'Enter your email to receive a 6-digit one-time login code.'
              : `We sent a 6-digit code to ${email}`}
          </p>
        </div>

        {error && (
          <div className="p-3 rounded-2xl bg-pink/10 border border-pink/30 text-pink text-xs text-center">
            {error}
          </div>
        )}

        {step === 'email' ? (
          <form onSubmit={handleRequestOtp} className="space-y-4">
            <div className="relative">
              <Mail className="w-4 h-4 text-orange absolute left-4 top-1/2 -translate-y-1/2" />
              <input
                type="email"
                required
                placeholder="you@domain.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full bg-dark border border-zinc-800 focus:border-orange rounded-2xl pl-11 pr-4 py-3 text-sm text-white placeholder:text-zinc-500 focus:outline-none"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-pink hover:opacity-95 text-white font-outfit font-bold text-sm rounded-2xl shadow-[0_0_20px_rgba(255,51,181,0.4)] transition-all flex items-center justify-center space-x-2 disabled:opacity-50"
            >
              {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <ArrowRight className="w-4 h-4" />}
              <span>{loading ? 'Sending Code...' : 'Send Login Code'}</span>
            </button>
          </form>
        ) : (
          <form onSubmit={handleVerifyOtp} className="space-y-4">
            <div className="relative">
              <KeyRound className="w-4 h-4 text-pink absolute left-4 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                inputMode="numeric"
                maxLength={6}
                required
                placeholder="000000"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
                className="w-full bg-dark border border-zinc-800 focus:border-pink rounded-2xl pl-11 pr-4 py-3 text-center font-mono text-lg tracking-[0.4em] text-orange placeholder:text-zinc-600 focus:outline-none"
              />
            </div>

            <button
              type="submit"
              disabled={loading || code.length !== 6}
              className="w-full py-3 bg-pink hover:opacity-95 text-white font-outfit font-bold text-sm rounded-2xl shadow-[0_0_20px_rgba(255,51,181,0.4)] transition-all flex items-center justify-center space-x-2 disabled:opacity-50"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              <span>{loading ? 'Verifying...' : 'Verify & Open Library'}</span>
            </button>

            <button
              type="button"
              onClick={() => {
                setStep('email');
                setCode('');
                setError(null);
              }}
              className="w-full py-2 text-xs text-zinc-400 hover:text-white flex items-center justify-center space-x-1.5"
            >
              <RotateCcw className="w-3.5 h-3.5" />
              <span>Use a different email</span>
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default Page;