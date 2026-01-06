import { useState } from 'react';
import { 
    GetProject, 
    ListFiles, 
    ReadFile, 
    WriteFile, 
    DeleteFile, 
    CreateFile, 
    CreateDirectory 
} from '../../wailsjs/go/main/App';
import { Project, FileTreeNode } from '../types';

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
            console.log('ListFiles called with path:', path);
            const result = await ListFiles(path);
            console.log('ListFiles result:', result);
            if (result && result.files) {
                setFiles(result.files);
            } else {
                console.warn('No files returned from ListFiles');
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
        try {
            await CreateFile(path, content);
            if (project && project.path) {
                await loadFiles(project.path);
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
        try {
            await CreateDirectory(path);
            if (project && project.path) {
                await loadFiles(project.path);
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
        try {
            await DeleteFile(path);
            if (project && project.path) {
                await loadFiles(project.path);
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
        if (!expandedFolders.has(path)) {
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

    const reset = () => {
        setProject(null);
        setFiles([]);
        setCurrentPath('');
        setSelectedFile(null);
        setFileContent('');
        setExpandedFolders(new Set());
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
        setFileContent,
        loadProject,
        selectFile,
        saveFile,
        createFile,
        createDirectory,
        deleteFile,
        toggleFolder,
        reset
    };
};

