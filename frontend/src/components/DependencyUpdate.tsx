import { useEffect, useState } from 'react';
import { Loader2, CheckCircle2, XCircle, RefreshCw } from 'lucide-react';
import { Button } from './ui/button';
import { Progress } from './ui/progress';
import { CheckDependencies, SetupDependencies } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import bytoLogo from 'figma:asset/e1c6c4d1df3cefc4435d7cc603c42e22f058f10f.png';

interface DependencyUpdateProps {
  onComplete: () => void;
}

export function DependencyUpdate({ onComplete }: DependencyUpdateProps) {
  const [phase, setPhase] = useState<'checking' | 'updating' | 'complete' | 'error'>('checking');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [currentPackage, setCurrentPackage] = useState<string | null>(null);
  const [downloadProgress, setDownloadProgress] = useState(0);

  useEffect(() => {
    const runBootstrap = async () => {
      setPhase('checking');
      setErrorMessage(null);
      setCurrentPackage(null);
      setDownloadProgress(0);

      try {
        await CheckDependencies();
        setPhase('updating');
        await SetupDependencies();
      } catch (err) {
        console.error('Dependency bootstrap error:', err);
        setPhase('error');
        setErrorMessage(err instanceof Error ? err.message : String(err));
      }
    };

    runBootstrap();
  }, []);

  useEffect(() => {
    EventsOn(
      'dependency_progress',
      (data: { Name?: string; name?: string; Percentage?: number; percentage?: number }) => {
        const depName = data.Name ?? data.name ?? 'unknown';
        const pct = data.Percentage ?? data.percentage ?? 0;
        setCurrentPackage(depName);
        setDownloadProgress(pct);
      },
    );

    EventsOn('dependency_bootstrap_complete', () => {
      setPhase('complete');
      setCurrentPackage(null);
      setDownloadProgress(0);
      setTimeout(() => onComplete(), 600);
    });

    EventsOn(
      'dependency_bootstrap_failed',
      (data: { error?: string }) => {
        setPhase('error');
        setErrorMessage(data?.error || 'Unknown error');
        setCurrentPackage(null);
        setDownloadProgress(0);
      },
    );

    return () => {
      EventsOff('dependency_progress');
      EventsOff('dependency_bootstrap_complete');
      EventsOff('dependency_bootstrap_failed');
    };
  }, [onComplete]);

  const handleRetry = () => {
    setPhase('checking');
    setErrorMessage(null);
    setCurrentPackage(null);
    setDownloadProgress(0);
    SetupDependencies();
  };

  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-[#0a0a0a]">
      <div className="flex flex-col items-center max-w-md w-full mx-auto px-6 space-y-8">
        <img src={bytoLogo} alt="Byto" className="size-12 opacity-90" />
        <h1 className="text-xl font-medium text-gray-100">Dependency Update</h1>

        {phase === 'checking' && (
          <div className="flex flex-col items-center gap-3 w-full">
            <Loader2 className="size-8 text-blue-400 animate-spin" />
            <p className="text-sm text-gray-500">Checking dependencies...</p>
          </div>
        )}

        {phase === 'updating' && (
          <div className="flex flex-col items-center gap-4 w-full">
            <p className="text-sm text-gray-500 text-center">
              Downloading necessary files
            </p>
            <div className="w-full space-y-2">
              {currentPackage ? (
                <>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-200 font-medium">{currentPackage}</span>
                    <span className="text-blue-400">{downloadProgress}%</span>
                  </div>
                  <Progress value={downloadProgress} className="h-2" />
                </>
              ) : (
                <div className="flex items-center gap-2">
                  <Loader2 className="size-4 text-blue-400 animate-spin" />
                  <span className="text-sm text-gray-500">Preparing...</span>
                </div>
              )}
            </div>
          </div>
        )}

        {phase === 'complete' && (
          <div className="flex flex-col items-center gap-3">
            <CheckCircle2 className="size-12 text-green-400 animate-in fade-in duration-300" />
            <p className="text-sm text-gray-500">All dependencies are up to date.</p>
          </div>
        )}

        {phase === 'error' && (
          <div className="w-full space-y-3">
            <p className="text-sm text-red-400 text-center">{errorMessage}</p>
            <Button onClick={handleRetry} className="w-full">
              <RefreshCw className="size-4 mr-2" />
              Retry
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
