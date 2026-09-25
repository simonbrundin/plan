import type { Goal, GoalDependency } from "~/types/goal";

interface GoalWithWeight extends Goal {
	weight: number;
	order: number;
}

interface GoalData {
	goal: Goal;
	children: GoalWithWeight[];
	parents: Goal[];
	allGoals: Goal[];
	dependencies: GoalDependency[];
	dependsOn: Goal[];
	blocking: Goal[];
	statusHistory?: StatusUpdate[];
}

interface StatusUpdate {
	id: number;
	goal_id: number;
	from_status_id?: number;
	to_status_id: number;
	changed_at: string;
	from_status_name?: string;
	to_status_name: string;
}

// Custom error for authentication failures
export class AuthenticationError extends Error {
	constructor(message = "Authentication required") {
		super(message);
		this.name = "AuthenticationError";
	}
}

// Token info stored in session
interface TokenInfo {
	accessToken: string;
	refreshToken?: string;
	expiresAt?: number;
}

export function useGoalApi() {
	const config = useRuntimeConfig();
	const { user, loggedIn, fetch: refreshSession } = useUserSession();
	const goApiUrl = config.public.goApiUrl || "http://localhost:8080";
	
	// Get token info from session
	const getTokenInfo = (): TokenInfo => {
		const userData = user.value as any;
		return {
			accessToken: userData?.accessToken,
			refreshToken: userData?.refreshToken,
			expiresAt: userData?.expiresAt,
		};
	};
	
	// Check if token needs refresh (expired or will expire in 5 minutes)
	const isTokenExpiringSoon = (): boolean => {
		const tokenInfo = getTokenInfo();
		if (!tokenInfo.accessToken) return true;
		if (!tokenInfo.expiresAt) return false; // No expiry info, assume valid
		// Refresh if expires within 5 minutes
		return Date.now() >= (tokenInfo.expiresAt - 5 * 60 * 1000);
	};

	// Refresh the access token
	const refreshAccessToken = async (): Promise<boolean> => {
		try {
			const response = await $fetch<{ success: boolean; expiresAt: number }>('/api/auth/refresh', {
				method: 'POST',
			});
			// Refresh the client-side session
			await refreshSession();
			console.log('Token refreshed successfully');
			return true;
		} catch (error) {
			console.error('Failed to refresh token:', error);
			return false;
		}
	};

	const authHeaders = () => {
		// Check if user is logged in
		if (!loggedIn.value || !user.value) {
			throw new AuthenticationError("Du måste vara inloggad för att utföra denna åtgärd");
		}
		
		const tokenInfo = getTokenInfo();
		
		if (!tokenInfo.accessToken) {
			throw new AuthenticationError("Ingen giltig session hittades. Vänligen logga in igen.");
		}
		
		return {
			Authorization: `Bearer ${tokenInfo.accessToken}`,
		};
	};

	// Make API request with automatic token refresh on 401
	const apiRequest = async <T>(url: string, options: RequestInit = {}): Promise<T> => {
		const response = await $fetch<T>(url, {
			...options,
			headers: {
				...options.headers,
				...authHeaders(),
			},
		});
		return response;
	};

	// API request that handles 401 and retries with refreshed token
	const apiRequestWithRetry = async <T>(
		url: string, 
		options: RequestInit = {},
		retryOn401 = true
	): Promise<T> => {
		try {
			return await apiRequest<T>(url, options);
		} catch (error: any) {
			// If 401 and we haven't retried yet, try refreshing token
			if (retryOn401 && error?.status === 401) {
				console.log('Got 401, attempting token refresh...');
				const refreshed = await refreshAccessToken();
				if (refreshed) {
					// Retry the request with new token
					return await apiRequest<T>(url, options);
				}
			}
			throw error;
		}
	};

	const fetchGoalData = async (
		goalId: number,
		forceRefresh = false,
	): Promise<GoalData> => {
		const url = forceRefresh
			? `${goApiUrl}/goals/${goalId}?_=${Date.now()}`
			: `${goApiUrl}/goals/${goalId}`;
		return await apiRequestWithRetry<GoalData>(url);
	};

	const updateGoalTitle = async (goalId: number, title: string) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { title },
		});
	};

	const updateGoalIcon = async (goalId: number, icon: string) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { icon },
		});
	};

	const updateGoalStatus = async (goalId: number, statusId: number) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/${goalId}/status`, {
			method: "PATCH",
			body: { status_id: statusId },
		});
	};

	const toggleGoalStarted = async (goalId: number, started: string | null) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { started },
		});
	};

	const toggleGoalFinished = async (
		goalId: number,
		finished: string | null,
	) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/${goalId}`, {
			method: "PATCH",
			body: { finished },
		});
	};

	const deleteGoal = async (goalId: number) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/${goalId}`, {
			method: "DELETE",
		});
	};

	const addParentRelation = async (childId: number, parentId: number) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId },
		});
	};

	const removeParentRelation = async (childId: number, parentId: number) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/relations`, {
			method: "DELETE",
			body: { childId, parentId },
		});
	};

	const addChildRelation = async (
		childId: number,
		parentId: number,
		order: number,
	) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId, order },
		});
	};

	const updateGoalOrder = async (
		parentId: number,
		childId: number,
		order: number,
	) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, order },
		});
	};

	const updateGoalWeight = async (
		parentId: number,
		childId: number,
		weight: number,
	) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/relations`, {
			method: "PATCH",
			body: { childId, parentId, weight },
		});
	};

	const setGoalWeight = async (
		childId: number,
		parentId: number,
		weight: number,
	) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/relations`, {
			method: "POST",
			body: { childId, parentId, weight },
		});
	};

	const loadAllGoals = async (): Promise<Goal[]> => {
		return await apiRequestWithRetry<Goal[]>(`${goApiUrl}/goals`);
	};

	const createGoal = async (
		title: string,
		statusId?: number,
	): Promise<Goal> => {
		const body: { title: string; status_id?: number } = { title };
		if (statusId) {
			body.status_id = statusId;
		}
		return await apiRequestWithRetry<Goal>(`${goApiUrl}/goals`, {
			method: "POST",
			body,
		});
	};

	const addDependency = async (
		goalId: number,
		dependsOnId: number,
	): Promise<GoalDependency> => {
		return await apiRequestWithRetry<GoalDependency>(`${goApiUrl}/goals/dependencies`, {
			method: "POST",
			body: { goalId, dependsOnId },
		});
	};

	const removeDependency = async (goalId: number, dependsOnId: number) => {
		return await apiRequestWithRetry(`${goApiUrl}/goals/dependencies`, {
			method: "DELETE",
			body: { goalId, dependsOnId },
		});
	};

	return {
		fetchGoalData,
		updateGoalTitle,
		updateGoalIcon,
		updateGoalStatus,
		toggleGoalStarted,
		toggleGoalFinished,
		deleteGoal,
		addParentRelation,
		removeParentRelation,
		addChildRelation,
		updateGoalOrder,
		updateGoalWeight,
		setGoalWeight,
		createGoal,
		loadAllGoals,
		addDependency,
		removeDependency,
		refreshAccessToken,
	};
}
