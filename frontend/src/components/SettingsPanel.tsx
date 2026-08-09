import { useState, useEffect } from 'react';
import { RefreshCw, Download, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { Button } from './ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { GetAppVersion, PerformFullUpdate, DownloadAppUpdate, LaunchInstaller } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

interface SettingsPanelProps {
  parallelDownloads: string;
  onClose: () => void;
  onSave: (settings: { parallelDownloads: string }) => void;
  initialUpdateState?: UpdateStatusState;
}

type UpdateStatus = 'idle' | 'checking' | 'deps-updating' | 'app-checking' | 'update-available' | 'downloading' | 'ready-to-install' | 'done' | 'error';

export interface UpdateStatusState {
  status: UpdateStatus;
  message: string;
  depsResult?: { success: boolean; message: string };
  appResult?: {
    success: boolean;
    message: string;
    current_version: string;
    latest_version: string;
    has_update: boolean;
    changelog: string;
    download_url: string;
  };
  downloadProgress?: number;
  installerPath?: string;
}

export function SettingsPanel({
  parallelDownloads: initialParallelDownloads,
  onClose,
  onSave,
  initialUpdateState
}: SettingsPanelProps) {
  // Local state - only applied when Save is clicked
  const [localParallelDownloads, setLocalParallelDownloads] = useState(initialParallelDownloads);

  // Update state
  const [appVersion, setAppVersion] = useState('');
  const [updateState, setUpdateState] = useState<UpdateStatusState>(initialUpdateState || {
    status: 'idle',
    message: ''
  });
  const [changelogOpen, setChangelogOpen] = useState(false);

  // Load app version on mount
  useEffect(() => {
    GetAppVersion().then(setAppVersion).catch(console.error);
  }, []);

  // Listen for update events
  useEffect(() => {
    const unsubStatus = EventsOn('update_status', (data: { step: string; message: string }) => {
      if (data.step === 'deps') {
        setUpdateState(prev => ({ ...prev, status: 'deps-updating', message: data.message }));
      } else if (data.step === 'app_check') {
        setUpdateState(prev => ({ ...prev, status: 'app-checking', message: data.message }));
      }
    });

    const unsubProgress = EventsOn('update_download_progress', (data: { downloaded: number; total: number; percentage: number }) => {
      setUpdateState(prev => ({
        ...prev,
        status: 'downloading',
        message: `Downloading update... ${data.percentage.toFixed(1)}%`,
        downloadProgress: data.percentage
      }));
    });

    return () => {
      EventsOff('update_status');
      EventsOff('update_download_progress');
    };
  }, []);

  const handleSave = () => {
    onSave({
      parallelDownloads: localParallelDownloads
    });
  };

  const handleCheckForUpdates = async () => {
    setUpdateState({ status: 'checking', message: 'Starting update check...' });

    try {
      const result = await PerformFullUpdate();

      const depsResult = result.deps as { success: boolean; message: string };
      const appResult = result.app as {
        success: boolean;
        message: string;
        current_version: string;
        latest_version: string;
        has_update: boolean;
        changelog: string;
        download_url: string;
      };

      if (appResult.has_update) {
        setUpdateState({
          status: 'update-available',
          message: `New version available: ${appResult.latest_version}`,
          depsResult,
          appResult
        });
      } else {
        setUpdateState({
          status: 'done',
          message: 'Everything is up to date!',
          depsResult,
          appResult
        });
      }
    } catch (error) {
      setUpdateState({
        status: 'error',
        message: `Update check failed: ${error}`
      });
    }
  };

  const handleDownloadUpdate = async () => {
    if (!updateState.appResult?.download_url) return;

    setUpdateState(prev => ({
      ...prev,
      status: 'downloading',
      message: 'Starting download...',
      downloadProgress: 0
    }));

    try {
      const installerPath = await DownloadAppUpdate(updateState.appResult.download_url);
      setUpdateState(prev => ({
        ...prev,
        status: 'ready-to-install',
        message: 'Download complete! Ready to install.',
        installerPath
      }));
    } catch (error) {
      setUpdateState(prev => ({
        ...prev,
        status: 'error',
        message: `Download failed: ${error}`
      }));
    }
  };

  const handleInstallUpdate = async () => {
    if (!updateState.installerPath) return;

    try {
      await LaunchInstaller(updateState.installerPath);
      // The app will close after launching the installer
    } catch (error) {
      setUpdateState(prev => ({
        ...prev,
        status: 'error',
        message: `Failed to launch installer: ${error}`
      }));
    }
  };

  const getStatusIcon = () => {
    switch (updateState.status) {
      case 'checking':
      case 'deps-updating':
      case 'app-checking':
      case 'downloading':
        return <Loader2 className="size-4 animate-spin text-blue-400" />;
      case 'done':
        return <CheckCircle className="size-4 text-green-400" />;
      case 'update-available':
      case 'ready-to-install':
        return <Download className="size-4 text-yellow-400" />;
      case 'error':
        return <AlertCircle className="size-4 text-red-400" />;
      default:
        return null;
    }
  };

  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl bg-[#141414] border border-[#262626] text-gray-100">
        <DialogHeader>
          <DialogTitle className="text-gray-100">Settings</DialogTitle>
          <DialogDescription className="text-gray-400">
            Configure your download preferences
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Updates Section */}
          <div className="p-4 bg-[#1a1a1a] rounded-lg border border-[#262626]">
            <div className="flex items-center justify-between mb-3">
              <div>
                <label className="text-gray-300 text-sm font-medium">Updates</label>
                <p className="text-gray-500 text-xs">Current version: {appVersion || 'Loading...'}</p>
              </div>
              <Button
                size="sm"
                variant="outline"
                className="border-[#262626] hover:bg-[#1f1f1f]"
                onClick={handleCheckForUpdates}
                disabled={['checking', 'deps-updating', 'app-checking', 'downloading'].includes(updateState.status)}
              >
                <RefreshCw className={`size-4 mr-2 ${['checking', 'deps-updating', 'app-checking'].includes(updateState.status) ? 'animate-spin' : ''}`} />
                Check for Updates
              </Button>
            </div>

            {/* Update Status */}
            {updateState.status !== 'idle' && (
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm">
                  {getStatusIcon()}
                  <span className={
                    updateState.status === 'error' ? 'text-red-400' :
                      updateState.status === 'done' ? 'text-green-400' :
                        updateState.status === 'update-available' ? 'text-yellow-400' :
                          'text-gray-300'
                  }>
                    {updateState.message}
                  </span>
                </div>

                {/* Dependencies result */}
                {updateState.depsResult && (
                  <div className="text-xs text-gray-400 pl-6">
                    <span className={updateState.depsResult.success ? 'text-green-400' : 'text-red-400'}>
                      Dependencies: {updateState.depsResult.message}
                    </span>
                  </div>
                )}

                {/* Download progress bar */}
                {updateState.status === 'downloading' && updateState.downloadProgress !== undefined && (
                  <div className="w-full bg-[#262626] rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${updateState.downloadProgress}%` }}
                    />
                  </div>
                )}

                {/* Changelog link */}
                {updateState.appResult?.has_update && updateState.appResult.changelog && (
                  <div>
                    <button
                      onClick={() => setChangelogOpen(true)}
                      className="text-xs text-blue-400 hover:text-blue-300 underline cursor-pointer pl-6"
                    >
                      What's new in v{updateState.appResult.latest_version}
                    </button>
                  </div>
                )}

                {/* Action buttons */}
                {updateState.status === 'update-available' && (
                  <Button
                    size="sm"
                    className="bg-blue-600 hover:bg-blue-700"
                    onClick={handleDownloadUpdate}
                  >
                    <Download className="size-4 mr-2" />
                    Download Update
                  </Button>
                )}

                {updateState.status === 'ready-to-install' && (
                  <Button
                    size="sm"
                    className="!bg-green-600 !hover:bg-green-700"
                    onClick={handleInstallUpdate}
                  >
                    <RefreshCw className="size-4 mr-2" />
                    Install & Restart
                  </Button>
                )}
              </div>
            )}
          </div>

          <div>
            <label className="text-gray-300 text-sm">Parallel Downloads</label>
            <p className="text-gray-500 text-xs mb-2">Number of simultaneous downloads</p>
            <Select
              value={localParallelDownloads}
              onValueChange={(val) => setLocalParallelDownloads(val)}
            >
              <SelectTrigger className="w-full bg-[#1f1f1f] border-[#262626] text-gray-100 h-9">
                <SelectValue placeholder="Select parallel downloads" />
              </SelectTrigger>
              <SelectContent className="bg-[#141414] border-[#262626] text-gray-100">
                <SelectItem value="1">1 (Sequential)</SelectItem>
                <SelectItem value="2">2</SelectItem>
                <SelectItem value="3">3</SelectItem>
                <SelectItem value="4">4</SelectItem>
                <SelectItem value="5">5</SelectItem>
                <SelectItem value="10">10</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-4 border-t border-[#262626]">
          <Button variant="outline" onClick={onClose} className="border-[#262626] hover:bg-[#1f1f1f]">
            Cancel
          </Button>
          <Button onClick={handleSave} className="bg-blue-600 hover:bg-blue-700">
            Save Changes
          </Button>
        </div>
      </DialogContent>

      {/* Changelog Dialog */}
      <Dialog open={changelogOpen} onOpenChange={setChangelogOpen}>
        <DialogContent className="max-w-lg max-h-[80vh] bg-[#141414] border border-[#262626] text-gray-100 flex flex-col">
          <DialogHeader>
            <DialogTitle className="text-gray-100">
              What's new in v{updateState.appResult?.latest_version}
            </DialogTitle>
            <DialogDescription className="text-gray-400">
              Changelog
            </DialogDescription>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto pr-2">
            <pre className="text-sm text-gray-300 whitespace-pre-wrap break-words font-sans">
              {updateState.appResult?.changelog}
            </pre>
          </div>
        </DialogContent>
      </Dialog>
    </Dialog>
  );
}