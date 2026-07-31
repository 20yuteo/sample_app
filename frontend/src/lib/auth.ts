const AUTH_ISSUER = process.env.NEXT_PUBLIC_AUTH_ISSUER ?? "http://localhost:18080/realms/commerce";
const AUTH_CLIENT_ID = process.env.NEXT_PUBLIC_AUTH_CLIENT_ID ?? "commerce-frontend";

export type AuthProfile = {
  email?: string;
  name?: string;
  preferred_username?: string;
  realm_access?: {
    roles?: string[];
  };
};

export type AuthSession = {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
  profile: AuthProfile;
};

const verifierKey = "commerce.auth.pkce.verifier";
const stateKey = "commerce.auth.state";
const clientIDKey = "commerce.auth.client_id";
const sessionKey = "commerce.auth.session";

export function getStoredSession(): AuthSession | null {
  if (typeof window === "undefined") {
    return null;
  }
  const raw = window.localStorage.getItem(sessionKey);
  if (!raw) {
    return null;
  }
  const session = JSON.parse(raw) as AuthSession;
  if (session.expiresAt <= Date.now()) {
    clearSession();
    return null;
  }
  return session;
}

export function storeSession(session: AuthSession) {
  window.localStorage.setItem(sessionKey, JSON.stringify(session));
}

export function clearSession() {
  window.localStorage.removeItem(sessionKey);
}

export async function startLogin(clientId = AUTH_CLIENT_ID) {
  const state = crypto.randomUUID();
  const verifier = randomString(64);
  const challenge = await sha256Base64URL(verifier);
  window.sessionStorage.setItem(stateKey, state);
  window.sessionStorage.setItem(verifierKey, verifier);
  window.sessionStorage.setItem(clientIDKey, clientId);

  const url = new URL(`${AUTH_ISSUER}/protocol/openid-connect/auth`);
  url.searchParams.set("client_id", clientId);
  url.searchParams.set("redirect_uri", callbackURL());
  url.searchParams.set("response_type", "code");
  url.searchParams.set("scope", "openid profile email roles");
  url.searchParams.set("state", state);
  url.searchParams.set("code_challenge", challenge);
  url.searchParams.set("code_challenge_method", "S256");
  window.location.assign(url.toString());
}

export async function completeLogin(code: string, state: string) {
  const expectedState = window.sessionStorage.getItem(stateKey);
  const verifier = window.sessionStorage.getItem(verifierKey);
  const clientId = window.sessionStorage.getItem(clientIDKey) ?? AUTH_CLIENT_ID;
  if (!expectedState || !verifier || expectedState !== state) {
    throw new Error("ログイン状態を検証できませんでした。もう一度ログインしてください。");
  }

  const body = new URLSearchParams({
    grant_type: "authorization_code",
    client_id: clientId,
    redirect_uri: callbackURL(),
    code,
    code_verifier: verifier
  });

  const response = await fetch(`${AUTH_ISSUER}/protocol/openid-connect/token`, {
    method: "POST",
    headers: {
      "content-type": "application/x-www-form-urlencoded"
    },
    body
  });
  if (!response.ok) {
    throw new Error("トークンを取得できませんでした。");
  }

  const token = (await response.json()) as {
    access_token: string;
    refresh_token?: string;
    expires_in: number;
  };
  const profile = parseJwt(token.access_token);
  const session: AuthSession = {
    accessToken: token.access_token,
    refreshToken: token.refresh_token,
    expiresAt: Date.now() + token.expires_in * 1000,
    profile
  };
  storeSession(session);
  window.sessionStorage.removeItem(stateKey);
  window.sessionStorage.removeItem(verifierKey);
  window.sessionStorage.removeItem(clientIDKey);
  return session;
}

export function logoutURL() {
  const url = new URL(`${AUTH_ISSUER}/protocol/openid-connect/logout`);
  url.searchParams.set("post_logout_redirect_uri", window.location.origin);
  clearSession();
  return url.toString();
}

function callbackURL() {
  return `${window.location.origin}/auth/callback`;
}

function parseJwt(token: string): AuthProfile {
  const [, payload] = token.split(".");
  if (!payload) {
    return {};
  }
  return JSON.parse(new TextDecoder().decode(base64URLToBytes(payload))) as AuthProfile;
}

function randomString(length: number) {
  const bytes = new Uint8Array(length);
  crypto.getRandomValues(bytes);
  return base64URL(bytes).slice(0, length);
}

async function sha256Base64URL(value: string) {
  const bytes = new TextEncoder().encode(value);
  const digest = await crypto.subtle.digest("SHA-256", bytes);
  return base64URL(new Uint8Array(digest));
}

function base64URL(bytes: Uint8Array) {
  return btoa(String.fromCharCode(...bytes)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function base64URLToBytes(value: string) {
  const padded = value.replace(/-/g, "+").replace(/_/g, "/").padEnd(Math.ceil(value.length / 4) * 4, "=");
  const raw = atob(padded);
  return Uint8Array.from(raw, (char) => char.charCodeAt(0));
}
