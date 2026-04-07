import { useState } from 'react';
import { Star, Github, Coffee, Smartphone, Copy, Check } from 'lucide-react';
import { Button } from './ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

interface SupportPanelProps {
  onClose: () => void;
}

export function SupportPanel({ onClose }: SupportPanelProps) {
  const [showVCash, setShowVCash] = useState(false);
  const [copied, setCopied] = useState(false);

  const openLink = (url: string) => {
    BrowserOpenURL(url);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText('+201006311537');
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <>
      <Dialog open={true} onOpenChange={(open: boolean) => {
        if (!open) onClose();
      }}>
        <DialogContent className="max-w-md bg-[#141414] border border-[#262626] text-gray-100">
          <DialogHeader>
            <DialogTitle className="text-gray-100">Support Byto</DialogTitle>
            <DialogDescription className="text-gray-400">
              If you find Byto useful, consider supporting the project
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3 py-4">
            <Button
              variant="outline"
              className="w-full justify-start gap-2 border-[#262626] hover:bg-[#1f1f1f]"
              onClick={() => setShowVCash(true)}
            >
              <Smartphone className="size-4" />
              Support via Vodafone Cash
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start gap-2 border-[#262626] hover:bg-[#1f1f1f]"
              onClick={() => openLink('https://ko-fi.com/omarnaru')}
            >
              <Coffee className="size-4" />
              Support me on Ko-fi
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start gap-2 border-[#262626] hover:bg-[#1f1f1f]"
              onClick={() => openLink('https://github.com/OmarNaru1110')}
            >
              <Github className="size-4" />
              Follow me on GitHub
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start gap-2 border-[#262626] hover:bg-[#1f1f1f]"
              onClick={() => openLink('https://github.com/OmarNaru1110/byto')}
            >
              <Star className="size-4" />
              Star/Contribute to byto on GitHub
            </Button>
          </div>

          <div className="flex justify-end pt-4 border-t border-[#262626]">
            <Button onClick={onClose} className="bg-blue-600 hover:bg-blue-700">
              Close
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={showVCash} onOpenChange={setShowVCash}>
        <DialogContent className="max-w-md bg-[#141414] border border-[#262626] text-gray-100">
          <DialogHeader>
            <DialogTitle className="text-gray-100 flex items-center gap-2">
              <Smartphone className="size-5 text-red-500" />
              Vodafone Cash
            </DialogTitle>
            <DialogDescription className="text-gray-400">
              You can send your support to the following number:
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col items-center justify-center py-6 space-y-6">
            <div className="flex items-center gap-3 px-6 py-4 bg-[#1f1f1f] rounded-xl border border-[#262626] w-full justify-center shadow-inner">
              <span className="text-3xl font-mono tracking-wider font-bold text-gray-100">
                +20 100 631 1537
              </span>
            </div>
            <Button
              variant="outline"
              onClick={handleCopy}
              className="w-full gap-2 border-[#262626] hover:bg-[#2a2a2a] py-6 text-lg transition-colors"
            >
              {copied ? (
                <>
                  <Check className="size-5 text-green-500" />
                  Copied!
                </>
              ) : (
                <>
                  <Copy className="size-5" />
                  Copy Number
                </>
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
