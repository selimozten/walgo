import { useState, useEffect } from 'react';
import { ListProjects } from '../../wailsjs/go/main/App';
import { Project } from '../types';

export const useProjects = () => {
    const [projects, setProjects] = useState<Project[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const loadProjects = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await ListProjects();
            if (result) {
                const mappedProjects = result.map((project) => ({
                    ...project,
                    deployments: project.deployCount || 0,
                    deploymentHistory: Array.isArray(project.deployments)
                        ? project.deployments.map((d) => ({
                            timestamp: d.createdAt,
                            objectId: d.objectId,
                            network: d.network,
                            epochs: d.epochs,
                            gasFee: d.gasFee,
                            status: d.success ? 'success' as const : 'failed' as const,
                        }))
                        : []
                }));
                setProjects(mappedProjects as Project[]);
            }
        } catch (err) {
            console.error('Failed to load projects:', err);
            setError('Failed to load projects');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadProjects();
    }, []);

    return {
        projects,
        loading,
        error,
        reloadProjects: loadProjects,
        setProjects
    };
};

