import React from 'react';
import { Folder, File, ChevronRight, ChevronDown, Trash2 } from 'lucide-react';
import { cn } from '../../utils/helpers';
import { FileTreeNode } from '../../types';

interface TreeNodeProps {
    node: FileTreeNode;
    level?: number;
    expandedFolders: Set<string>;
    toggleFolder: (path: string) => void;
    selectFile: (file: FileTreeNode) => void;
    deleteSelectedFile: (file: FileTreeNode) => void;
    selectedFile: FileTreeNode | null;
}

export const TreeNode: React.FC<TreeNodeProps> = ({ 
    node, 
    level = 0, 
    expandedFolders, 
    toggleFolder, 
    selectFile, 
    deleteSelectedFile,
    selectedFile 
}) => {
    const isExpanded = expandedFolders.has(node.path);
    const isSelected = selectedFile?.path === node.path;
    const indent = level * 16;

    return (
        <div key={node.path}>
            <div
                className={cn(
                    "flex items-center gap-2 px-3 py-2 cursor-pointer transition-all hover:bg-white/5 group",
                    isSelected && "bg-accent/10 border border-accent/30"
                )}
                style={{ paddingLeft: `${indent + 12}px` }}
                onClick={() => {
                    if (node.isDir) {
                        toggleFolder(node.path);
                    } else {
                        selectFile(node);
                    }
                }}
            >
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
                {node.isDir ? (
                    <Folder size={16} className={cn(
                        "flex-shrink-0",
                        isExpanded ? "text-accent" : "text-zinc-500 group-hover:text-accent"
                    )} />
                ) : (
                    <File size={16} className={cn(
                        "flex-shrink-0",
                        node.name.endsWith('.md') ? "text-blue-400" : "text-zinc-500",
                        isSelected && "text-accent"
                    )} />
                )}
                <span className={cn(
                    "flex-1 truncate text-xs font-mono transition-colors",
                    isSelected ? "text-accent" : "text-zinc-300 group-hover:text-white"
                )}>
                    {node.name}
                </span>
                <button
                    type="button"
                    onClick={(e) => {
                        e.stopPropagation();
                        e.preventDefault();
                        deleteSelectedFile(node);
                    }}
                    onPointerDown={(e) => {
                        e.stopPropagation();
                        e.preventDefault();
                    }}
                    className="opacity-50 hover:opacity-100 hover:bg-red-500/20 rounded-sm p-2 z-10 relative transition-all"
                    title={`Delete ${node.isDir ? 'directory' : 'file'}`}
                >
                    <Trash2 size={14} className="text-red-400" />
                </button>
            </div>
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
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

