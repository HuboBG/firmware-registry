import axios from "axios";
import { runtimeConfig } from "./runtime-config";
import { getAccessToken } from "./auth";
import { logger } from "./logger";

const api = axios.create({
    baseURL: runtimeConfig.API_BASE_URL || "",
    timeout: 20000
});

// Add response interceptor for logging
api.interceptors.response.use(
    (response) => {
        logger.debug('API request successful', {
            method: response.config.method,
            url: response.config.url,
            status: response.status,
        });
        return response;
    },
    (error) => {
        logger.error('API request failed', error, {
            method: error.config?.method,
            url: error.config?.url,
            status: error.response?.status,
            statusText: error.response?.statusText,
        });
        return Promise.reject(error);
    }
);

const adminHeaders = () => {
    // If OIDC is enabled, use Bearer token
    if (runtimeConfig.OIDC_ENABLED) {
        const token = getAccessToken();
        return token ? { "Authorization": `Bearer ${token}` } : {};
    }
    // Fall back to API key
    const k = runtimeConfig.ADMIN_KEY;
    return k ? {"X-Admin-Key": k} : {};
};

const deviceHeaders = () => {
    // If OIDC is enabled, use Bearer token
    if (runtimeConfig.OIDC_ENABLED) {
        const token = getAccessToken();
        return token ? { "Authorization": `Bearer ${token}` } : {};
    }
    // Fall back to API key
    const k = runtimeConfig.DEVICE_KEY;
    return k ? {"X-Device-Key": k} : {};
};

export interface FirmwareDTO {
    type: string;
    version: string;
    filename: string;
    sizeBytes: number;
    sha256: string;
    createdAt: string;
    downloadUrl?: string;
}

export interface WebhookDTO {
    id: number;
    url: string;
    events: string[];
    enabled: boolean;
}

export const FirmwareAPI = {
    async list(type: string): Promise<FirmwareDTO[]> {
        const r = await api.get(`/api/firmware/${type}`, {headers: deviceHeaders()});
        return r.data as FirmwareDTO[];
    },

    async latest(type: string): Promise<FirmwareDTO> {
        const r = await api.get(`/api/firmware/${type}/latest`, {headers: deviceHeaders()});
        return r.data as FirmwareDTO;
    },

    async upload(type: string, version: string, file: File): Promise<FirmwareDTO> {
        const fd = new FormData();
        fd.append("file", file);
        const r = await api.post(`/api/firmware/${type}/${version}`, fd, {
            headers: {...adminHeaders(), "Content-Type": "multipart/form-data"}
        });
        return r.data as FirmwareDTO;
    },

    async remove(type: string, version: string): Promise<{ deleted: boolean }> {
        const r = await api.delete(`/api/firmware/${type}/${version}`, {headers: adminHeaders()});
        return r.data as { deleted: boolean };
    }
};

export const WebhookAPI = {
    async list(): Promise<WebhookDTO[]> {
        const r = await api.get(`/api/webhooks`, {headers: adminHeaders()});
        return r.data as WebhookDTO[];
    },

    async create(hook: Omit<WebhookDTO, "id">): Promise<{ id: number }> {
        const r = await api.post(`/api/webhooks`, hook, {headers: adminHeaders()});
        return r.data as { id: number };
    },

    async update(id: number, hook: Omit<WebhookDTO, "id">): Promise<{ updated: boolean }> {
        const r = await api.put(`/api/webhooks/${id}`, hook, {headers: adminHeaders()});
        return r.data as { updated: boolean };
    },

    async remove(id: number): Promise<{ deleted: boolean }> {
        const r = await api.delete(`/api/webhooks/${id}`, {headers: adminHeaders()});
        return r.data as { deleted: boolean };
    }
};
