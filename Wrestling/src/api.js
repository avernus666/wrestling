const API_BASE_URL = process.env.REACT_APP_API_URL || "/api";

const request = async (path, signal) => {
  const response = await fetch(`${API_BASE_URL}${path}`, { signal });
  if (!response.ok) {
    throw new Error(`Ошибка API: ${response.status}`);
  }
  return response.json();
};

export const fetchElements = (signal) => request("/elements", signal);
export const fetchBars = (signal) => request("/bars", signal);
export const fetchBase = (signal) => request("/base", signal);
export const fetchSafety = (signal) => request("/safety", signal);
export const fetchStats = (signal) => request("/stats", signal);

const authRequest = async (path, options = {}) => {
  const token = localStorage.getItem("wrestling-token");
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}), ...(options.headers || {}) },
  });
  const body = response.status === 204 ? null : await response.json().catch(() => null);
  if (!response.ok) throw new Error(body?.error || `Ошибка API: ${response.status}`);
  return body;
};
export const register = (payload) => authRequest("/auth/register", { method: "POST", body: JSON.stringify(payload) });
export const login = (payload) => authRequest("/auth/login", { method: "POST", body: JSON.stringify(payload) });
export const me = () => authRequest("/auth/me");
export const logout = () => authRequest("/auth/logout", { method: "POST" });
export const fetchProgress = () => authRequest("/progress");
export const saveProgress = (payload) => authRequest("/progress", { method: "PUT", body: JSON.stringify(payload) });
export const createWorkoutSession = (payload) => authRequest("/workouts", { method: "POST", body: JSON.stringify(payload) });
export const completeWorkoutSession = (id, completedIDs) => authRequest(`/workouts/${id}`, { method: "POST", body: JSON.stringify({ completedIds: completedIDs }) });
export const fetchHistory = () => authRequest("/history");
export const fetchAnalytics = () => authRequest("/analytics");
export const fetchSkills = () => authRequest("/skills");
export const saveSkill = (payload) => authRequest("/skills", { method: "PUT", body: JSON.stringify(payload) });
export const fetchCommunity = (category = "", search = "") => request(`/community?category=${encodeURIComponent(category)}&search=${encodeURIComponent(search)}`);
export const createCommunity = (payload) => authRequest("/community", { method: "POST", body: JSON.stringify(payload) });
export const fetchCommunityComments = (id) => request(`/community/${encodeURIComponent(id)}/comments`);
export const createCommunityComment = (id, body) => authRequest(`/community/${encodeURIComponent(id)}/comments`, { method: "POST", body: JSON.stringify({ body }) });

export const reportCommunity = (id, reason) => authRequest(`/community/${encodeURIComponent(id)}/report`, { method: "POST", body: JSON.stringify({ reason }) });
