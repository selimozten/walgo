import React, { useState, useEffect } from 'react';
import { Folder, File, ChevronRight, ChevronDown } from 'lucide-react';
import { cn } from '../../utils/helpers';
import { FileTreeNode } from '../../types';
import { ContextMenu } from './ContextMenu';

interface TreeNodeProps {
    node: FileTreeNode;
    level?: number;
    expandedFolders: Set<string>;
    toggleFolder: (path: string) => void;
    selectFile: (file: FileTreeNode) => void;
    deleteSelectedFile: (file: FileTreeNode) => void;
    selectedFile: FileTreeNode | null;
    onRename?: (node: FileTreeNode, newName: string) => Promise<void>;
    onMove?: (sourcePath: string, targetPath: string) => Promise<void>;
    onCopy?: (node: FileTreeNode) => void;
    onCut?: (node: FileTreeNode) => void;
    onPaste?: (targetNode: FileTreeNode) => Promise<void>;
    onDuplicate?: (node: FileTreeNode) => void;
    onNewFile?: (parentNode: FileTreeNode) => void;
    onNewFolder?: (parentNode: FileTreeNode) => void;
    clipboard?: { node: FileTreeNode; operation: 'copy' | 'cut' } | null;
    checkDepth?: (node: FileTreeNode) => Promise<boolean>;
}

export const TreeNode: React.FC<TreeNodeProps> = ({
    node,
    level = 0,
    expandedFolders,
    toggleFolder,
    selectFile,
    deleteSelectedFile,
    selectedFile,
    onRename,
    onMove,
    onCopy,
    onCut,
    onPaste,
    onDuplicate,
    onNewFile,
    onNewFolder,
    clipboard,
    checkDepth,
}) => {
    const isExpanded = expandedFolders.has(node.path);
    const isSelected = selectedFile?.path === node.path;
    const indent = level * 16;

    // State for context menu
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null);
    const [isTooDeep, setIsTooDeep] = useState(false);

    // State for rename mode
    const [isRenaming, setIsRenaming] = useState(false);
    const [renameName, setRenameName] = useState(node.name);

    // State for drag and drop
    const [isDragging, setIsDragging] = useState(false);
    const [isDragOver, setIsDragOver] = useState(false);

    // Check if this node is in clipboard (cut operation)
    const isCut = clipboard?.operation === 'cut' && clipboard?.node.path === node.path;

    // Cleanup drag-over state when component unmounts or path changes
    useEffect(() => {
        return () => {
            setIsDragOver(false);
        };
    }, [node.path]);

    const handleContextMenu = async (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();

        // Check depth if this is a directory
        if (node.isDir && checkDepth) {
            const tooDeep = await checkDepth(node);
            setIsTooDeep(tooDeep);
        }

        setContextMenu({
            x: e.clientX,
            y: e.clientY,
        });
    };

    const handleClick = () => {
        if (contextMenu) {
            setContextMenu(null);
        }
        if (!isRenaming) {
            if (node.isDir) {
                toggleFolder(node.path);
            } else {
                selectFile(node);
            }
        }
    };

    const handleRenameStart = () => {
        setIsRenaming(true);
        setRenameName(node.name);
    };

    const handleRenameConfirm = async () => {
        if (renameName && renameName !== node.name && onRename) {
            await onRename(node, renameName);
        }
        setIsRenaming(false);
    };

    const handleRenameCancel = () => {
        setIsRenaming(false);
        setRenameName(node.name);
    };

    // Drag and drop handlers
    const handleDragStart = (e: React.DragEvent) => {
        e.stopPropagation();
        setIsDragging(true);
        e.dataTransfer.effectAllowed = 'move';
        e.dataTransfer.setData('application/json', JSON.stringify({
            path: node.path,
            name: node.name,
            isDir: node.isDir
        }));
    };

    const handleDragEnd = () => {
        setIsDragging(false);
    };

    const handleDragOver = (e: React.DragEvent) => {
        if (!node.isDir) return;

        e.preventDefault();
        e.stopPropagation();
        setIsDragOver(true);
        e.dataTransfer.dropEffect = 'move';
    };

    const handleDragLeave = (e: React.DragEvent) => {
        // Check if we're actually leaving this element (not just entering a child)
        const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
        const isLeavingElement =
            e.clientX < rect.left ||
            e.clientX >= rect.right ||
            e.clientY < rect.top ||
            e.clientY >= rect.bottom;

        if (isLeavingElement) {
            setIsDragOver(false);
        }
    };

    const handleDrop = async (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragOver(false);

        if (!node.isDir || !onMove) return;

        try {
            const data = JSON.parse(e.dataTransfer.getData('application/json'));
            const sourcePath = data.path;
            const targetPath = `${node.path}/${data.name}`;

            // Don't drop on itself or its children
            if (sourcePath === node.path || targetPath.startsWith(sourcePath + '/')) {
                return;
            }

            // Check if source is already in this folder
            const sourceParent = sourcePath.substring(0, sourcePath.lastIndexOf('/'));
            if (sourceParent === node.path) {
                return; // Already in this folder
            }

            await onMove(sourcePath, targetPath);
        } catch (err) {
            console.error('Drop failed:', err);
        }
    };

    return (
        <div key={node.path}>
            <div
                className={cn(
                    "flex items-center gap-2 px-3 py-2 cursor-pointer transition-all relative group",
                    isSelected && "bg-accent/10 border border-accent/30",
                    !isSelected && "hover:bg-white/5",
                    isDragging && "opacity-40",
                    isDragOver && node.isDir && "bg-accent/20 border-2 border-accent/50",
                    isCut && "opacity-50"
                )}
                style={{ paddingLeft: `${indent + 12}px` }}
                onClick={handleClick}
                onContextMenu={handleContextMenu}
                draggable={!isRenaming}
                onDragStart={handleDragStart}
                onDragEnd={handleDragEnd}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
            >
                {/* Expand/collapse icon for directories */}
                {node.isDir && (
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            toggleFolder(node.path);
                        }}
                        className="flex-shrink-0 p-0 hover:bg-accent/20 rounded-sm transition-all"
                    >
                        {isExpanded ? (
                            <ChevronDown size={14} className="text-zinc-400" />
                        ) : (
                            <ChevronRight size={14} className="text-zinc-400" />
                        )}
                    </button>
                )}

                {/* File/Folder icon */}
                {node.isDir ? (
                    <Folder size={16} className={cn(
                        "flex-shrink-0",
                        node.name === 'content'
                            ? "text-green-500"
                            : isExpanded
                                ? "text-accent"
                                : "text-zinc-500 group-hover:text-accent"
                    )} />
                ) : (
                    <File size={16} className={cn(
                        "flex-shrink-0",
                        node.name.endsWith('.md') ? "text-blue-400" : "text-zinc-500",
                        isSelected && "text-accent"
                    )} />
                )}

                {/* Name (editable in rename mode) */}
                {isRenaming ? (
                    <input
                        type="text"
                        value={renameName}
                        onChange={(e) => setRenameName(e.target.value)}
                        onBlur={handleRenameConfirm}
                        onKeyDown={(e) => {
                            e.stopPropagation();
                            if (e.key === 'Enter') {
                                handleRenameConfirm();
                            } else if (e.key === 'Escape') {
                                handleRenameCancel();
                            }
                        }}
                        onClick={(e) => e.stopPropagation()}
                        autoFocus
                        autoComplete="off"
                        autoCapitalize="off"
                        autoCorrect="off"
                        spellCheck={false}
                        className="flex-1 px-2 py-1 bg-zinc-800 border border-accent/50 rounded text-xs font-mono text-white focus:outline-none focus:border-accent"
                    />
                ) : (
                    <span className={cn(
                        "flex-1 truncate text-xs font-mono transition-colors",
                        node.name === 'content' && node.isDir
                            ? "text-green-500 font-semibold"
                            : isSelected
                                ? "text-accent"
                                : "text-zinc-300 group-hover:text-white"
                    )}>
                        {node.name}
                    </span>
                )}
            </div>

            {/* Context Menu */}
            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    node={node}
                    onClose={() => setContextMenu(null)}
                    onDelete={deleteSelectedFile}
                    onRename={handleRenameStart}
                    onCopy={() => onCopy?.(node)}
                    onCut={() => onCut?.(node)}
                    onPaste={onPaste ? () => onPaste(node) : undefined}
                    onDuplicate={() => onDuplicate?.(node)}
                    onNewFile={() => onNewFile?.(node)}
                    onNewFolder={() => onNewFolder?.(node)}
                    hasClipboard={clipboard !== null && clipboard !== undefined}
                    isTooDeep={isTooDeep}
                />
            )}

            {/* Render children if directory is expanded */}
            {node.isDir && isExpanded && node.children && node.children.length > 0 && (
                <div>
                    {node.children.map((child) => (
                        <TreeNode
                            key={child.path}
                            node={child}
                            level={level + 1}
                            expandedFolders={expandedFolders}
                            toggleFolder={toggleFolder}
                            selectFile={selectFile}
                            deleteSelectedFile={deleteSelectedFile}
                            selectedFile={selectedFile}
                            onRename={onRename}
                            onMove={onMove}
                            onCopy={onCopy}
                            onCut={onCut}
                            onPaste={onPaste}
                            onDuplicate={onDuplicate}
                            onNewFile={onNewFile}
                            onNewFolder={onNewFolder}
                            clipboard={clipboard}
                            checkDepth={checkDepth}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};
