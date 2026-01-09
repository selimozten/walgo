import React, { useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { Trash2, Edit3, Copy, FilePlus, FolderPlus, Scissors, Clipboard, Files } from 'lucide-react';
import { FileTreeNode } from '../../types';
import { cn } from '../../utils';

interface ContextMenuProps {
  x: number;
  y: number;
  node: FileTreeNode;
  onClose: () => void;
  onDelete?: (node: FileTreeNode) => void;
  onRename?: (node: FileTreeNode) => void;
  onCopy?: (node: FileTreeNode) => void;
  onCut?: (node: FileTreeNode) => void;
  onPaste?: (targetNode: FileTreeNode) => void;
  onDuplicate?: (node: FileTreeNode) => void;
  onNewFile?: (parentNode: FileTreeNode) => void;
  onNewFolder?: (parentNode: FileTreeNode) => void;
  hasClipboard?: boolean;
  isTooDeep?: boolean;
}

export const ContextMenu: React.FC<ContextMenuProps> = ({
  x,
  y,
  node,
  onClose,
  onDelete,
  onRename,
  onCopy,
  onCut,
  onPaste,
  onDuplicate,
  onNewFile,
  onNewFolder,
  hasClipboard = false,
  isTooDeep = false,
}) => {
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    const timer = setTimeout(() => {
      document.addEventListener('click', handleClick, true);
      document.addEventListener('contextmenu', handleClick, true);
      document.addEventListener('keydown', handleKeyDown);
    }, 10);

    return () => {
      clearTimeout(timer);
      document.removeEventListener('click', handleClick, true);
      document.removeEventListener('contextmenu', handleClick, true);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [onClose]);

  const getMenuPosition = () => {
    const menuWidth = 220;
    const menuHeight = 300;
    const padding = 10;
    const { innerWidth, innerHeight } = window;

    let left = x;
    let top = y;

    if (left + menuWidth > innerWidth - padding) {
      left = Math.max(padding, innerWidth - menuWidth - padding);
    }

    if (top + menuHeight > innerHeight - padding) {
      top = Math.max(padding, innerHeight - menuHeight - padding);
    }

    return { left, top };
  };

  const position = getMenuPosition();

  const MenuItem = ({
    icon: Icon,
    label,
    onClick,
    danger = false,
    shortcut
  }: {
    icon: React.ElementType;
    label: string;
    onClick: () => void;
    danger?: boolean;
    shortcut?: string;
  }) => (
    <button
      onClick={(e) => {
        e.stopPropagation();
        onClick();
        onClose();
      }}
      className={cn(
        "w-full px-3 py-2 text-left text-xs font-mono transition-all flex items-center justify-between gap-3 group",
        danger
          ? "text-red-400 hover:bg-red-500/10 hover:text-red-300"
          : "text-zinc-300 hover:bg-white/5 hover:text-white"
      )}
    >
      <div className="flex items-center gap-3">
        <Icon size={14} className={danger ? "" : "text-zinc-500 group-hover:text-accent"} />
        <span>{label}</span>
      </div>
      {shortcut && <span className="text-[10px] text-zinc-600">{shortcut}</span>}
    </button>
  );

  const menu = (
    <div
      ref={menuRef}
      className="fixed z-[99999] min-w-[200px] bg-zinc-900 border border-white/20 rounded-md shadow-2xl overflow-hidden"
      style={{ left: `${position.left}px`, top: `${position.top}px` }}
      onClick={(e) => e.stopPropagation()}
      onContextMenu={(e) => e.preventDefault()}
    >
      <div className="py-1">
        {node.isDir && (onNewFile || onNewFolder) && (
          <>
            {onNewFile && <MenuItem icon={FilePlus} label="New File" onClick={() => onNewFile(node)} />}
            {onNewFolder && <MenuItem icon={FolderPlus} label="New Folder" onClick={() => onNewFolder(node)} />}
            <div className="h-px bg-white/5 my-1" />
          </>
        )}

        {onRename && <MenuItem icon={Edit3} label="Rename" onClick={() => onRename(node)} shortcut="F2" />}
        {!isTooDeep && onDuplicate && <MenuItem icon={Files} label="Duplicate" onClick={() => onDuplicate(node)} shortcut="⌘D" />}
        {!isTooDeep && onCopy && <MenuItem icon={Copy} label="Copy" onClick={() => onCopy(node)} shortcut="⌘C" />}
        {!isTooDeep && onCut && <MenuItem icon={Scissors} label="Cut" onClick={() => onCut(node)} shortcut="⌘X" />}
        {isTooDeep && (onCopy || onCut || onDuplicate) && (
          <div className="px-3 py-2 text-xs font-mono text-yellow-500">
            Directory too deep for copy/cut
          </div>
        )}

        {node.isDir && hasClipboard && onPaste && (
          <MenuItem icon={Clipboard} label="Paste" onClick={() => onPaste(node)} shortcut="⌘V" />
        )}

        {(onRename || onDuplicate || onCopy || onCut || (node.isDir && hasClipboard && onPaste)) && onDelete && (
          <div className="h-px bg-white/5 my-1" />
        )}

        {onDelete && (
          <MenuItem
            icon={Trash2}
            label={node.isDir ? "Delete Folder" : "Delete File"}
            onClick={() => onDelete(node)}
            danger
            shortcut="Del"
          />
        )}
      </div>
    </div>
  );

  return createPortal(menu, document.body);
};
