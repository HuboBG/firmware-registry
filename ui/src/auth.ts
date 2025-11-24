import { UserManager, User, WebStorageStateStore } from 'oidc-client-ts';
import { ref, computed } from 'vue';
import { logger } from './logger';

export interface AuthConfig {
    enabled: boolean;
    authority: string;
    clientId: string;
    redirectUri: string;
    scope: string;
}

// Global reactive auth state
export const currentUser = ref<User | null>(null);
export const isAuthenticated = computed(() => currentUser.value !== null && !currentUser.value.expired);

let userManager: UserManager | null = null;

export function initAuth(config: AuthConfig) {
    if (!config.enabled) {
        logger.info('Authentication disabled, using legacy API keys');
        return;
    }

    logger.info('Initializing OIDC authentication', {
        authority: config.authority,
        clientId: config.clientId,
    });

    userManager = new UserManager({
        authority: config.authority,
        client_id: config.clientId,
        redirect_uri: config.redirectUri,
        response_type: 'code',
        scope: config.scope,
        post_logout_redirect_uri: config.redirectUri,
        userStore: new WebStorageStateStore({ store: window.localStorage }),
        automaticSilentRenew: true,
    });

    // Load existing user from storage
    userManager.getUser().then(user => {
        if (user && !user.expired) {
            currentUser.value = user;
            logger.info('User loaded from storage', {
                profile: user.profile,
            });
        }
    }).catch((error) => {
        logger.error('Failed to load user from storage', error);
    });

    // Handle silent renew success
    userManager.events.addUserLoaded((user) => {
        currentUser.value = user;
        logger.debug('Token silently renewed');
    });

    // Handle silent renew error
    userManager.events.addSilentRenewError((error) => {
        logger.error('Silent token renewal failed', error);
    });

    // Handle user signed out
    userManager.events.addUserSignedOut(() => {
        currentUser.value = null;
        logger.info('User signed out');
    });
}

export async function login() {
    if (!userManager) {
        const error = new Error('Auth not initialized');
        logger.error('Login failed: auth not initialized', error);
        throw error;
    }
    logger.info('Initiating login redirect');
    try {
        await userManager.signinRedirect();
    } catch (error) {
        logger.error('Login redirect failed', error as Error);
        throw error;
    }
}

export async function handleCallback() {
    if (!userManager) {
        const error = new Error('Auth not initialized');
        logger.error('Callback handling failed: auth not initialized', error);
        throw error;
    }
    logger.info('Handling authentication callback');
    try {
        const user = await userManager.signinRedirectCallback();
        currentUser.value = user;
        logger.info('User authenticated successfully', {
            profile: user.profile,
        });
        return user;
    } catch (error) {
        logger.error('Authentication callback failed', error as Error);
        throw error;
    }
}

export async function logout() {
    if (!userManager) {
        const error = new Error('Auth not initialized');
        logger.error('Logout failed: auth not initialized', error);
        throw error;
    }
    logger.info('Initiating logout');
    try {
        await userManager.signoutRedirect();
    } catch (error) {
        logger.error('Logout redirect failed', error as Error);
        throw error;
    }
}

export function getAccessToken(): string | null {
    return currentUser.value?.access_token || null;
}

export function getUserProfile() {
    return currentUser.value?.profile || null;
}
