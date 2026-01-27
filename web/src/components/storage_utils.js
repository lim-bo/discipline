const jwtTokenKey = "Auth-Token";

export const getToken = () => {
    return localStorage.getItem(jwtTokenKey);
}

export const setToken = (token) => {
    if (typeof(token) !== "string") {
        return;
    }
    localStorage.setItem(jwtTokenKey, token);
}

export const clearToken = () => {
    localStorage.removeItem(jwtTokenKey);
}