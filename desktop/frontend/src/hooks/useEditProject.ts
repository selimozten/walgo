import { useState } from 'react';
import {
    GetProject,
    ListFiles,
    ReadFile,
    WriteFile,
    DeleteFile,
    CreateFile,
    CreateDirectory,
    RenameFile,
    MoveFile,
    CopyFile
} from '../../wailsjs/go/main/App';
import { Project, FileTreeNode } from '../types';

// Sort comparator: directories first, then files, both alphabetically
const sortFilesComparator = (a: FileTreeNode, b: FileTreeNode) => {
    if (a.isDir === b.isDir) return a.name.localeCompare(b.name);
    return a.isDir ? -1 : 1;
};

export const useEditProject = () => {
    const [project, setProject] = useState<Project | null>(null);
    const [files, setFiles] = useState<FileTreeNode[]>([]);
    const [currentPath, setCurrentPath] = useState('');
    const [selectedFile, setSelectedFile] = useState<FileTreeNode | null>(null);
    const [fileContent, setFileContent] = useState('');
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set());
    const [clipboard, setClipboard] = useState<{ node: FileTreeNode; operation: 'copy' | 'cut' } | null>(null);

    const loadProject = async (projectInput: string | number | Project) => {
        setLoading(true);
        setError(null);
        try {
            let proj: Project;

            if (typeof projectInput === 'object') {
                // It's already a Project object
                proj = projectInput;
            } else if (typeof projectInput === 'number') {
                // It's a project ID, fetch from API
                const apiProject = await GetProject(projectInput);
                proj = {
                    name: apiProject.name,
                    path: apiProject.sitePath,
                    sitePath: apiProject.sitePath,
                };
            } else {
                // It's a project name - fetch from API to get the full path
                console.warn('Loading project by name is deprecated, please pass full project object');
                proj = {
                    name: projectInput,
                    path: undefined,
                };
            }

            setProject(proj);
            // Check both path and sitePath for compatibility
            const projectPath = proj.path || proj.sitePath;
            if (projectPath) {
                console.log('Loading files from path:', projectPath);
                await loadFiles(projectPath);
            } else {
                console.error('Project has no path or sitePath:', proj);
                setError('Project path is missing');
            }
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            console.error('Failed to load project:', errorMsg, err);
            setError(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    const loadFiles = async (path: string) => {
        setLoading(true);
        setError(null);
        try {
            const result = await ListFiles(path);
            if (result && result.files) {
                // Sort files: directories first, then files, both alphabetically
                const sortedFiles = [...result.files].sort(sortFilesComparator);
                setFiles(sortedFiles);
            } else {
                setFiles([]);
            }
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            console.error('Failed to load files:', errorMsg, err);
            setError(errorMsg);
            setFiles([]);
        } finally {
            setLoading(false);
        }
    };

    const selectFile = async (file: FileTreeNode) => {
        setSelectedFile(file);
        if (!file.isDir) {
            setLoading(true);
            try {
                const content = await ReadFile(file.path);
                setFileContent(content.content || '');
            } catch (err) {
                console.error('Failed to read file:', err);
            } finally {
                setLoading(false);
            }
        }
    };

    const saveFile = async () => {
        if (!selectedFile) return;
        
        setSaving(true);
        setError(null);
        try {
            await WriteFile(selectedFile.path, fileContent);
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setSaving(false);
        }
    };

    const createFile = async (path: string, content: string) => {
        setLoading(true);
        setError(null);

        // Preserve expanded folders state
        const previousExpandedFolders = new Set(expandedFolders);

        try {
            await CreateFile(path, content);
            const projectPath = project?.path || project?.sitePath;
            if (project && projectPath) {
                // Reload root files first
                const rootResult = await ListFiles(projectPath);
                if (rootResult && rootResult.files) {
                    const sortedFiles = [...rootResult.files].sort(sortFilesComparator);
                    setExpandedFolders(previousExpandedFolders);

                    // Reload children for all expanded folders with fresh root
                    await reloadExpandedFolders(previousExpandedFolders, sortedFiles);
                }
            }
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const createDirectory = async (path: string) => {
        setLoading(true);
        setError(null);

        // Preserve expanded folders state
        const previousExpandedFolders = new Set(expandedFolders);

        try {
            await CreateDirectory(path);
            const projectPath = project?.path || project?.sitePath;
            if (project && projectPath) {
                // Reload root files first
                const rootResult = await ListFiles(projectPath);
                if (rootResult && rootResult.files) {
                    const sortedFiles = [...rootResult.files].sort(sortFilesComparator);
                    setExpandedFolders(previousExpandedFolders);

                    // Reload children for all expanded folders with fresh root
                    await reloadExpandedFolders(previousExpandedFolders, sortedFiles);
                }
            }
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const deleteFile = async (path: string) => {
        setLoading(true);
        setError(null);

        // Preserve expanded folders state
        const previousExpandedFolders = new Set(expandedFolders);

        try {
            await DeleteFile(path);
            const projectPath = project?.path || project?.sitePath;
            if (project && projectPath) {
                // Restore expanded folders, removing the deleted path and its children
                const newExpandedFolders = new Set<string>();
                previousExpandedFolders.forEach(folderPath => {
                    if (folderPath !== path && !folderPath.startsWith(path + '/')) {
                        newExpandedFolders.add(folderPath);
                    }
                });

                // Reload root files first
                const rootResult = await ListFiles(projectPath);
                if (rootResult && rootResult.files) {
                    const sortedFiles = [...rootResult.files].sort(sortFilesComparator);
                    setExpandedFolders(newExpandedFolders);

                    // Reload children for all expanded folders with fresh root
                    await reloadExpandedFolders(newExpandedFolders, sortedFiles);
                }
            }
            if (selectedFile && selectedFile.path === path) {
                setSelectedFile(null);
                setFileContent('');
            }
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const toggleFolder = async (path: string) => {
        const wasExpanded = expandedFolders.has(path);

        setExpandedFolders(prev => {
            const next = new Set(prev);
            if (next.has(path)) {
                next.delete(path);
            } else {
                next.add(path);
            }
            return next;
        });

        // Load children if folder is being expanded and doesn't have children yet
        if (!wasExpanded) {
            try {
                const result = await ListFiles(path);
                if (result && result.files && result.files.length > 0) {
                    // Update the files tree to include children
                    setFiles(prevFiles => {
                        const updateNode = (nodes: FileTreeNode[]): FileTreeNode[] => {
                            return nodes.map(node => {
                                if (node.path === path) {
                                    return {
                                        ...node,
                                        children: result.files as FileTreeNode[]
                                    };
                                } else if (node.children) {
                                    return {
                                        ...node,
                                        children: updateNode(node.children)
                                    };
                                }
                                return node;
                            });
                        };
                        return updateNode(prevFiles);
                    });
                }
            } catch (err) {
                console.error('Failed to load folder contents:', err);
            }
        }
    };

    const reloadExpandedFolders = async (expandedPaths: Set<string>, rootFiles: FileTreeNode[]) => {
        // Helper function to recursively load children for expanded folders
        const loadChildrenRecursively = async (nodes: FileTreeNode[]): Promise<FileTreeNode[]> => {
            const updatedNodes: FileTreeNode[] = [];

            for (const node of nodes) {
                if (node.isDir && expandedPaths.has(node.path)) {
                    // This folder is expanded, load its children
                    try {
                        const result = await ListFiles(node.path);
                        if (result && result.files) {
                            const sortedChildren = [...result.files].sort(sortFilesComparator);
                            // Recursively load children of expanded child folders
                            const childrenWithNested = await loadChildrenRecursively(sortedChildren);
                            updatedNodes.push({ ...node, children: childrenWithNested });
                        } else {
                            updatedNodes.push(node);
                        }
                    } catch (err) {
                        console.error('Failed to load children for', node.path, err);
                        updatedNodes.push(node);
                    }
                } else {
                    updatedNodes.push(node);
                }
            }

            return updatedNodes;
        };

        // Start from provided root files
        const updatedFiles = await loadChildrenRecursively(rootFiles);
        setFiles(updatedFiles);
    };

    const renameFile = async (node: FileTreeNode, newName: string) => {
        setLoading(true);
        setError(null);

        // Preserve expanded folders state
        const previousExpandedFolders = new Set(expandedFolders);

        try {
            const result = await RenameFile(node.path, newName);
            if (!result.success) {
                throw new Error(result.error || 'Rename failed');
            }

            // Update expanded folders, updating the renamed folder's path if it was a directory
            let updatedExpandedFolders = previousExpandedFolders;
            if (node.isDir) {
                const newExpandedFolders = new Set<string>();
                previousExpandedFolders.forEach(path => {
                    if (path === node.path) {
                        // Update to new path
                        const parentPath = node.path.substring(0, node.path.lastIndexOf('/'));
                        newExpandedFolders.add(`${parentPath}/${newName}`);
                    } else if (path.startsWith(node.path + '/')) {
                        // Update child paths
                        const relativePath = path.substring(node.path.length);
                        const parentPath = node.path.substring(0, node.path.lastIndexOf('/'));
                        newExpandedFolders.add(`${parentPath}/${newName}${relativePath}`);
                    } else {
                        newExpandedFolders.add(path);
                    }
                });
                updatedExpandedFolders = newExpandedFolders;
            }

            setExpandedFolders(updatedExpandedFolders);

            // Update the name in the tree without full reload
            const updateNodeName = (nodes: FileTreeNode[]): FileTreeNode[] => {
                return nodes.map(n => {
                    if (n.path === node.path) {
                        const parentPath = node.path.substring(0, node.path.lastIndexOf('/'));
                        return { ...n, name: newName, path: `${parentPath}/${newName}` };
                    } else if (n.children) {
                        return { ...n, children: updateNodeName(n.children) };
                    }
                    return n;
                });
            };

            setFiles(prevFiles => updateNodeName(prevFiles));

            // Update selected file if it was renamed
            if (selectedFile && selectedFile.path === node.path) {
                setSelectedFile(null);
                setFileContent('');
            }
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const moveFile = async (sourcePath: string, targetPath: string) => {
        setLoading(true);
        setError(null);

        // Preserve expanded folders state
        const previousExpandedFolders = new Set(expandedFolders);

        try {
            const result = await MoveFile(sourcePath, targetPath);
            if (!result.success) {
                throw new Error(result.error || 'Move failed');
            }
            const projectPath = project?.path || project?.sitePath;
            if (project && projectPath) {
                // Reload root files first
                const rootResult = await ListFiles(projectPath);
                if (rootResult && rootResult.files) {
                    const sortedFiles = [...rootResult.files].sort(sortFilesComparator);
                    setExpandedFolders(previousExpandedFolders);

                    // Reload children for all expanded folders with fresh root
                    await reloadExpandedFolders(previousExpandedFolders, sortedFiles);
                }
            }
            // Clear selected file if it was moved
            if (selectedFile && selectedFile.path === sourcePath) {
                setSelectedFile(null);
                setFileContent('');
            }
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const copyFile = async (sourcePath: string, targetPath: string) => {
        setLoading(true);
        setError(null);

        // Preserve expanded folders state
        const previousExpandedFolders = new Set(expandedFolders);

        try {
            const result = await CopyFile(sourcePath, targetPath);
            if (!result.success) {
                throw new Error(result.error || 'Copy failed');
            }
            const projectPath = project?.path || project?.sitePath;
            if (project && projectPath) {
                // Reload root files first
                const rootResult = await ListFiles(projectPath);
                if (rootResult && rootResult.files) {
                    const sortedFiles = [...rootResult.files].sort(sortFilesComparator);
                    setExpandedFolders(previousExpandedFolders);

                    // Reload children for all expanded folders with fresh root
                    await reloadExpandedFolders(previousExpandedFolders, sortedFiles);
                }
            }
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const copyToClipboard = (node: FileTreeNode) => {
        setClipboard({ node, operation: 'copy' });
    };

    const cutToClipboard = (node: FileTreeNode) => {
        setClipboard({ node, operation: 'cut' });
    };

    const pasteFromClipboard = async (targetFolder: FileTreeNode) => {
        if (!clipboard) return { success: false, error: 'Nothing to paste' };

        const targetPath = `${targetFolder.path}/${clipboard.node.name}`;

        // Check if source and destination are the same
        if (clipboard.node.path === targetPath) {
            if (clipboard.operation === 'cut') {
                // For cut operation, just clear clipboard (file is already at destination)
                setClipboard(null);
                return { success: true };
            }
            // For copy operation, let it proceed (backend will auto-increment the name)
        }

        if (clipboard.operation === 'copy') {
            return await copyFile(clipboard.node.path, targetPath);
        } else {
            const result = await moveFile(clipboard.node.path, targetPath);
            if (result.success) {
                setClipboard(null); // Clear clipboard after cut operation
            }
            return result;
        }
    };

    const reset = () => {
        setProject(null);
        setFiles([]);
        setCurrentPath('');
        setSelectedFile(null);
        setFileContent('');
        setExpandedFolders(new Set());
        setClipboard(null);
        setError(null);
    };

    return {
        project,
        files,
        currentPath,
        selectedFile,
        fileContent,
        loading,
        saving,
        error,
        expandedFolders,
        clipboard,
        setFileContent,
        setExpandedFolders,
        loadProject,
        selectFile,
        saveFile,
        createFile,
        createDirectory,
        deleteFile,
        renameFile,
        moveFile,
        copyToClipboard,
        cutToClipboard,
        pasteFromClipboard,
        toggleFolder,
        reset
    };
};

