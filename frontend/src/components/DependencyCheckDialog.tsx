import { useCallback, useEffect, useRef, useState, type KeyboardEvent } from 'react';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Button } from './ui/button';
import { CheckDependencies, SetupDependencies } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import type { deps } from '../../wailsjs/go/models';

interface Dependency {
  name: string;
  status: 'checking' | 'found' | 'missing' | 'downloading';
  version?: string;
  progress?: number;
}

interface DependencyCheckDialogProps {
  onClose: () => void;
}

export function DependencyCheckDialog({ onClose }: DependencyCheckDialogProps) {
  const [dependencies, setDependencies] = useState<Dependency[]>([
    { name: 'yt-dlp', status: 'checking' },
    { name: 'ffmpeg', status: 'checking' },
    { name: 'deno', status: 'checking' },
  ]);
  const hasAutoSetupRunRef = useRef(false);

  const DEP_ORDER = ['yt-dlp', 'ffmpeg', 'deno'] as const;

  const mapStateToDependency = (name: string, s: deps.DependencyState | undefined): Dependency => {
    if (!s) {
      return { name, status: 'missing' };
    }
    const exists = !!s.current_version && s.current_version.trim() !== '';
    if (exists) {
      return { name, status: 'found', version: s.current_version || undefined };
    }
    return { name, status: 'missing' };
  };

  const checkDependencies = useCallback(async () => {
    try {
      const states = await CheckDependencies();
      const byName = new Map(states.map((st) => [st.name, st]));
      let allReady = true;
      setDependencies(
        DEP_ORDER.map((name) => {
          const dep = mapStateToDependency(name, byName.get(name));
          if (dep.status !== 'found') allReady = false;
          return dep;
        }),
      );
      if (allReady) {
        setTimeout(() => onClose(), 500);
      }
    } catch (err) {
      console.error('Failed to check dependencies:', err);
      setDependencies(DEP_ORDER.map((name) => ({ name, status: 'missing' as const })));
    }
  }, [onClose]);

  useEffect(() => {
    checkDependencies();
  }, [checkDependencies]);

  useEffect(() => {
    EventsOn(
      'dependency_progress',
      (data: { Name?: string; name?: string; Percentage?: number; percentage?: number }) => {
        const depName = data.Name ?? data.name ?? '';
        const pct = data.Percentage ?? data.percentage ?? 0;
        if (!depName) return;
        setDependencies((prev) =>
          prev.map((dep) =>
            dep.name === depName ? { ...dep, status: 'downloading', progress: Math.round(pct) } : dep,
          ),
        );
      },
    );
    return () => {
      EventsOff('dependency_progress');
    };
  }, []);

  const setupMissingDependencies = useCallback(
    async () => {
      if (hasAutoSetupRunRef.current) {
        return;
      }

      hasAutoSetupRunRef.current = true;

      setDependencies((prev) =>
        prev.map((dep) =>
          dep.status === 'missing' ? { ...dep, status: 'downloading', progress: dep.progress ?? 0 } : dep,
        ),
      );

      try {
        await SetupDependencies();
      } catch (err) {
        console.error('Failed to set up dependencies:', err);
      } finally {
        await checkDependencies();
      }
    },
    [checkDependencies],
  );

  useEffect(() => {
    const hasMissing = dependencies.some((dep) => dep.status === 'missing');
    const isBusy = dependencies.some((dep) => dep.status === 'checking' || dep.status === 'downloading');

    if (hasMissing && !isBusy && !hasAutoSetupRunRef.current) {
      void setupMissingDependencies();
    }
  }, [dependencies, setupMissingDependencies]);

  const handleManualRetry = useCallback(() => {
    hasAutoSetupRunRef.current = false;
    void setupMissingDependencies();
  }, [setupMissingDependencies]);

  const getStatusIcon = (status: Dependency['status']) => {
    switch (status) {
      case 'checking':
        return <Loader2 className="size-4 text-blue-400 animate-spin" />;
      case 'found':
        return <CheckCircle2 className="size-4 text-green-400" />;
      case 'downloading':
        return <Loader2 className="size-4 text-blue-400 animate-spin" />;
      case 'missing':
        return <XCircle className="size-4 text-red-400" />;
    }
  };

  const getStatusText = (dep: Dependency) => {
    switch (dep.status) {
      case 'checking':
        return 'Checking...';
      case 'found':
        return dep.version ? `Found (${dep.version})` : 'Found';
      case 'downloading':
        return `${dep.progress || 0}%`;
      case 'missing':
        return 'Not found';
    }
  };

  // Only allow closing when all dependencies are found
  const canClose = dependencies.every((dep) => dep.status === 'found');

  return (
    <Dialog open={true} onOpenChange={(open: boolean) => !open && canClose && onClose()}>
      <DialogContent
        className="max-w-md bg-[#141414] border border-[#262626] text-gray-100"
        hideCloseButton={!canClose}
        onEscapeKeyDown={(e: KeyboardEvent) => {
          if (!canClose) e.preventDefault();
        }}
        onPointerDownOutside={(e: { preventDefault: () => void }) => {
          if (!canClose) e.preventDefault();
        }}
        onInteractOutside={(e: { preventDefault: () => void }) => {
          if (!canClose) e.preventDefault();
        }}
      >
        <DialogHeader>
          <DialogTitle className="text-gray-100">Checking Dependencies</DialogTitle>
        </DialogHeader>

        <div className="space-y-3 py-4">
          {dependencies.map((dep) => (
            <div key={dep.name} className="flex items-center gap-3 p-3 bg-[#0a0a0a] border border-[#262626] rounded-lg">
              {getStatusIcon(dep.status)}
              <span className="text-gray-200">{dep.name}</span>
              <span className="ml-auto text-sm text-gray-400">{getStatusText(dep)}</span>
              {dep.status === 'downloading' && (
                <div className="ml-2 w-16 bg-[#262626] rounded-full h-2 overflow-hidden">
                  <div
                    className="bg-blue-400 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${dep.progress || 0}%` }}
                  />
                </div>
              )}
            </div>
          ))}
        </div>

        {dependencies.some(dep => dep.status === 'missing') && hasAutoSetupRunRef.current && (
          <div className="pt-4 border-t border-[#262626] flex justify-end">
            <Button onClick={handleManualRetry} className="bg-blue-600 hover:bg-blue-700">
              Retry Download
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
