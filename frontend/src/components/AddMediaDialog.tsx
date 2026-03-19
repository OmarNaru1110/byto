import { useState, useEffect, CSSProperties } from 'react';
import { FolderOpen } from 'lucide-react';
import { Button } from './ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog';
import { SelectDownloadFolderWithDefault, GetMediaDefaults, UpdateMediaDefaults, SaveMediaDefaults, AddToQueue } from '../../wailsjs/go/main/App';
import { domain } from '../../wailsjs/go/models';

// Map backend quality (number) to frontend string
const qualityFromBackend: Record<number, string> = {
    0: '360p',
    1: '480p',
    2: '720p',
    3: '1080p',
    4: '1440p',
    5: '2160p',
};

interface AddMediaDialogProps {
    url: string;
    open: boolean;
    onClose: () => void;
    onSuccess: (id: string, quality: string, path: string) => void;
}

// --- Focusable input that shows a blue ring on focus ---
interface StyledInputProps {
    type?: string;
    min?: string;
    value: string;
    onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
    onClick?: (e: React.MouseEvent) => void;
    placeholder?: string;
    /** background colour — matches the card hover state */
    bg?: string;
}

function StyledInput({ type = 'text', min, value, onChange, onClick, placeholder, bg = '#0d0d0d' }: StyledInputProps) {
    const [focused, setFocused] = useState(false);

    const style: CSSProperties = {
        width: '100%',
        padding: '8px 12px',
        background: bg,
        border: `1px solid ${focused ? '#3b82f6' : '#2d2d2d'}`,
        borderRadius: '10px',
        fontSize: '13px',
        color: '#f3f4f6',
        outline: 'none',
        boxSizing: 'border-box',
        boxShadow: focused ? '0 0 0 3px rgba(59,130,246,0.15)' : 'none',
        transition: 'border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease',
    };

    return (
        <input
            type={type}
            min={min}
            value={value}
            onChange={onChange}
            onClick={onClick}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            style={style}
            placeholder={placeholder}
        />
    );
}

// --- Reusable styled option card — exposes hovered to children via render prop ---
interface OptionCardProps {
    isSelected: boolean;
    onClick: () => void;
    children: (hovered: boolean) => React.ReactNode;
}

function OptionCard({ isSelected, onClick, children }: OptionCardProps) {
    const [hovered, setHovered] = useState(false);

    const style: CSSProperties = {
        padding: '12px',
        borderRadius: '8px',
        border: `1px solid ${isSelected ? '#3b82f6' : hovered ? '#555' : '#2d2d2d'}`,
        background: isSelected
            ? 'rgba(59, 130, 246, 0.08)'
            : hovered
                ? '#212121'
                : '#181818',
        cursor: 'pointer',
        transition: 'border-color 0.15s ease, background 0.15s ease',
        outline: isSelected ? '1px solid rgba(59, 130, 246, 0.25)' : 'none',
        outlineOffset: '1px',
    };

    return (
        <div
            style={style}
            onClick={onClick}
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
        >
            {children(hovered)}
        </div>
    );
}

// --- Yes / No toggle button ---
interface ToggleButtonProps {
    isActive: boolean;
    onClick: () => void;
    children: React.ReactNode;
}

function ToggleButton({ isActive, onClick, children }: ToggleButtonProps) {
    const [hovered, setHovered] = useState(false);

    const style: CSSProperties = {
        flex: 1,
        padding: '8px 0',
        fontSize: '14px',
        fontWeight: 500,
        borderRadius: '6px',
        border: `1px solid ${isActive ? '#3b82f6' : hovered ? '#555' : '#2d2d2d'}`,
        background: isActive
            ? 'rgba(59, 130, 246, 0.1)'
            : hovered
                ? '#252525'
                : '#181818',
        color: isActive ? '#60a5fa' : hovered ? '#d1d5db' : '#9ca3af',
        cursor: 'pointer',
        transition: 'all 0.15s ease',
        boxShadow: isActive ? '0 0 12px rgba(59, 130, 246, 0.15)' : 'none',
    };

    return (
        <button
            style={style}
            onClick={onClick}
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
        >
            {children}
        </button>
    );
}

// --- Radio dot indicator ---
function RadioDot({ selected }: { selected: boolean }) {
    return (
        <div
            style={{
                width: '18px',
                height: '18px',
                borderRadius: '50%',
                border: `2px solid ${selected ? '#3b82f6' : '#4b5563'}`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
                transition: 'border-color 0.15s ease',
            }}
        >
            {selected && (
                <div
                    style={{
                        width: '8px',
                        height: '8px',
                        borderRadius: '50%',
                        background: '#3b82f6',
                    }}
                />
            )}
        </div>
    );
}

export function AddMediaDialog({ url, open, onClose, onSuccess }: AddMediaDialogProps) {
    const [quality, setQuality] = useState('1080p');
    const [downloadPath, setDownloadPath] = useState('');
    const [onlyAudio, setOnlyAudio] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    // Playlist state
    const [isPlaylist, setIsPlaylist] = useState(false);
    const [selectionType, setSelectionType] = useState('all'); // 'all', 'range', 'items'
    const [rangeStart, setRangeStart] = useState('1');
    const [rangeEnd, setRangeEnd] = useState('');
    const [specificItems, setSpecificItems] = useState('');

    // Load media defaults when dialog opens
    useEffect(() => {
        if (open) {
            loadDefaults();
            setIsPlaylist(false);
            setSelectionType('all');
            setRangeStart('1');
            setRangeEnd('');
            setSpecificItems('');
        }
    }, [open, url]);

    const loadDefaults = async () => {
        setIsLoading(true);
        try {
            const defaults = await GetMediaDefaults();
            if (defaults) {
                setQuality(qualityFromBackend[defaults.quality] || '1080p');
                setDownloadPath(defaults.download_path || '');
                setOnlyAudio(defaults.only_audio || false);
            }
        } catch (error) {
            console.error('Error loading media defaults:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleSelectFolder = async () => {
        try {
            const path = await SelectDownloadFolderWithDefault(downloadPath);
            if (path) setDownloadPath(path);
        } catch (error) {
            console.error('Error selecting folder:', error);
        }
    };

    const handleAdd = async () => {
        try {
            const selection = new domain.PlaylistSelection();
            selection.type = selectionType;
            if (selectionType === 'range') {
                selection.start_index = parseInt(rangeStart) || 1;
                selection.end_index = parseInt(rangeEnd) || parseInt(rangeStart) || 1;
            } else if (selectionType === 'items') {
                selection.items = specificItems;
            }
            const id = await AddToQueue(url, quality, downloadPath, onlyAudio, isPlaylist, selection);
            await UpdateMediaDefaults(quality, downloadPath, onlyAudio);
            await SaveMediaDefaults();
            onSuccess(id, quality, downloadPath);
            onClose();
        } catch (error) {
            console.error('Error adding to queue:', error);
        }
    };

    return (
        <Dialog open={open} onOpenChange={(isOpen: boolean) => !isOpen && onClose()}>
            {/*
              The DialogContent gets a max-h here so that the dialog itself is
              bounded. We then let the inner scrollable div reach the edge with
              negative-right-margin + matching right-padding so the scrollbar
              appears flush against the dialog border, not inside a padded box.
            */}
            <DialogContent
                className="max-w-lg bg-[#141414] border border-[#262626] text-gray-100"
                style={{ maxHeight: '90vh', display: 'flex', flexDirection: 'column' }}
            >
                <DialogHeader style={{ flexShrink: 0 }}>
                    <DialogTitle className="text-gray-100">Add Download</DialogTitle>
                    <DialogDescription className="text-gray-400">
                        Configure quality and download location for this media
                    </DialogDescription>
                </DialogHeader>

                {isLoading ? (
                    <div className="py-8 text-center text-gray-400">Loading...</div>
                ) : (
                    /* Scrollable body — negative right margin pushes scrollbar
                       to the dialog padding edge; padding-right compensates      */
                    <div
                        style={{
                            flex: 1,
                            overflowY: 'auto',
                            marginRight: '-24px',
                            paddingRight: '24px',
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '16px',
                            paddingTop: '16px',
                            paddingBottom: '16px',
                            colorScheme: 'dark',
                        }}
                    >
                        {/* Quality Selection */}
                        <div>
                            <label className="text-gray-300 text-sm font-medium block mb-1">Video Quality</label>
                            <p className="text-gray-500 text-xs mb-2">Select preferred quality for this download</p>
                            <select
                                value={quality}
                                onChange={(e) => setQuality(e.target.value)}
                                disabled={onlyAudio}
                                className="w-full px-3 py-2 bg-[#1f1f1f] border border-[#262626] rounded text-sm text-gray-100"
                                style={{ opacity: onlyAudio ? 0.5 : 1, cursor: onlyAudio ? 'not-allowed' : 'default' }}
                            >
                                <option value="360p">360p</option>
                                <option value="480p">480p</option>
                                <option value="720p">720p (HD)</option>
                                <option value="1080p">1080p (Full HD)</option>
                                <option value="1440p">1440p (2K)</option>
                                <option value="2160p">2160p (4K)</option>
                            </select>
                        </div>

                        {/* Playlist Options Section */}
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                            <label className="text-gray-300 text-sm font-medium block mb-1">Is this a playlist?</label>

                            {/* Yes / No Toggle */}
                            <div style={{ display: 'flex', gap: '10px'}}>
                                <ToggleButton isActive={isPlaylist} onClick={() => setIsPlaylist(true)}>
                                    Yes
                                </ToggleButton>
                                <ToggleButton isActive={!isPlaylist} onClick={() => setIsPlaylist(false)}>
                                    No (Single Video)
                                </ToggleButton>
                            </div>

                            {/* Download options — visible only when isPlaylist */}
                            {isPlaylist && (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                    <label className="text-gray-300 text-sm font-medium block mb-1">Download Options</label>

                                    {/* All Videos */}
                                    <OptionCard isSelected={selectionType === 'all'} onClick={() => setSelectionType('all')}>
                                        {() => (
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                                                <RadioDot selected={selectionType === 'all'} />
                                                <div>
                                                    <div style={{ fontSize: '14px', fontWeight: 500, color: selectionType === 'all' ? '#60a5fa' : '#d1d5db' }}>
                                                        Download All Videos
                                                    </div>
                                                    <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '2px' }}>
                                                        Download every video in the playlist
                                                    </div>
                                                </div>
                                            </div>
                                        )}
                                    </OptionCard>

                                    {/* Range */}
                                    <OptionCard isSelected={selectionType === 'range'} onClick={() => setSelectionType('range')}>
                                        {(hovered) => {
                                            // Darker when card is idle, slightly lighter when hovered/selected
                                            const inputBg = (selectionType === 'range' || hovered) ? '#151515' : '#0d0d0d';
                                            return (
                                                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '10px' }}>
                                                    <RadioDot selected={selectionType === 'range'} />
                                                    <div style={{ flex: 1, minWidth: 0 }}>
                                                        <div style={{ fontSize: '14px', fontWeight: 500, color: selectionType === 'range' ? '#60a5fa' : '#d1d5db' }}>
                                                            Download Range
                                                        </div>
                                                        <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '2px' }}>
                                                            Download videos from start to end position
                                                        </div>
                                                        {selectionType === 'range' && (
                                                            <div
                                                                style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '10px' }}
                                                                onClick={(e) => e.stopPropagation()}
                                                            >
                                                                <StyledInput
                                                                    type="number"
                                                                    min="1"
                                                                    value={rangeStart}
                                                                    onChange={(e) => setRangeStart(e.target.value)}
                                                                    placeholder="Start (1)"
                                                                    bg={inputBg}
                                                                />
                                                                <span style={{ color: '#4b5563', flexShrink: 0, fontSize: '16px' }}>→</span>
                                                                <StyledInput
                                                                    type="number"
                                                                    min="1"
                                                                    value={rangeEnd}
                                                                    onChange={(e) => setRangeEnd(e.target.value)}
                                                                    placeholder="End"
                                                                    bg={inputBg}
                                                                />
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            );
                                        }}
                                    </OptionCard>

                                    {/* Specific Items */}
                                    <OptionCard isSelected={selectionType === 'items'} onClick={() => setSelectionType('items')}>
                                        {(hovered) => {
                                            const inputBg = (selectionType === 'items' || hovered) ? '#151515' : '#0d0d0d';
                                            return (
                                                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '10px' }}>
                                                    <RadioDot selected={selectionType === 'items'} />
                                                    <div style={{ flex: 1, minWidth: 0 }}>
                                                        <div style={{ fontSize: '14px', fontWeight: 500, color: selectionType === 'items' ? '#60a5fa' : '#d1d5db' }}>
                                                            Specific Videos
                                                        </div>
                                                        <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '2px' }}>
                                                            Download specific videos by their position
                                                        </div>
                                                        {selectionType === 'items' && (
                                                            <div
                                                                style={{ marginTop: '10px' }}
                                                                onClick={(e) => e.stopPropagation()}
                                                            >
                                                                <StyledInput
                                                                    value={specificItems}
                                                                    onChange={(e) => setSpecificItems(e.target.value)}
                                                                    placeholder="e.g. 1,3,5,8"
                                                                    bg={inputBg}
                                                                />
                                                                <p style={{ fontSize: '11px', color: '#6b7280', marginTop: '4px' }}>
                                                                    Comma-separated positions: 1,3,5,8
                                                                </p>
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            );
                                        }}
                                    </OptionCard>
                                </div>
                            )}
                        </div>

                        {/* Audio Only Option */}
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                            <label className="text-gray-300 text-sm font-medium block mb-1">Format</label>
                            <OptionCard isSelected={onlyAudio} onClick={() => setOnlyAudio((prev) => !prev)}>
                                {() => (
                                    <div className="select-none" style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                                        <RadioDot selected={onlyAudio} />
                                        <div>
                                            <div style={{ fontSize: '14px', fontWeight: 500, color: onlyAudio ? '#60a5fa' : '#d1d5db' }}>
                                                Audio Only
                                            </div>
                                            <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '2px' }}>
                                                Extract audio and download it without downloading the whole media
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </OptionCard>
                        </div>

                        {/* Download Path */}
                        <div>
                            <label className="text-gray-300 text-sm font-medium block mb-1">Download Location</label>
                            <p className="text-gray-500 text-xs mb-2">Where this file will be saved</p>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    value={downloadPath}
                                    onChange={(e) => setDownloadPath(e.target.value)}
                                    className="flex-1 px-3 py-2 bg-[#1f1f1f] border border-[#262626] rounded text-sm text-gray-100"
                                />
                                <Button
                                    size="sm"
                                    variant="outline"
                                    className="border-[#262626] hover:bg-[#1f1f1f]"
                                    onClick={handleSelectFolder}
                                >
                                    <FolderOpen className="size-4" />
                                </Button>
                            </div>
                        </div>
                    </div>
                )}

                <DialogFooter
                    className="flex justify-end gap-2 pt-4 border-t border-[#262626]"
                    style={{ flexShrink: 0 }}
                >
                    <Button variant="outline" onClick={() => onClose()} className="border-[#262626] hover:bg-[#1f1f1f]">
                        Cancel
                    </Button>
                    <Button onClick={handleAdd} className="bg-blue-600 hover:bg-blue-700" disabled={isLoading || !downloadPath}>
                        Add to Queue
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
