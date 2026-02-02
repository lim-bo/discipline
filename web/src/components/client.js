export const apiConfig = {
    protocol: "http",
    address: "localhost:8070",
    version: "1",
    endpoints: {
        login: "/auth/login",
        register: "/auth/register",
        createHabit: "/habits",
        getHabits: "/habits",
        deleteHabit: "/habits",
        getChecks: "/habits/check"
    }
}

export const post = async (endpoint, token, reqBody) => {
    try {
        const headers = {
            "Content-Type": "application/json"
        };
        if (token) {
            headers.token = token;
        }
        const response = await fetch(
            `${apiConfig.protocol}://${apiConfig.address}/api/v${apiConfig.version}${endpoint}`,
            {
                method: "POST",
                body: JSON.stringify(reqBody),
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`
                }
            }
        );
        if (!response.ok) {
            throw response.status;
        }
        const body = await response.json();
        return body;
    } catch (e) {
        console.error(`Request error, statuscode: ${e}`);
        return { error: e };
    }    
}

export const get = async (endpoint, token) => {
    try {
        const response = await fetch(`${apiConfig.protocol}://${apiConfig.address}/api/v${apiConfig.version}${endpoint}`,
            {
                method: "GET",
                headers: {
                    "Authorization": `Bearer ${token}`
                }
            }
        );
        if (!response.ok) {
            throw response.status;
        }
        const body = await response.json();
        return body;
    } catch (e) {
        console.error(`Request error, statuscode: ${e}`);
        return {error: e};
    }
}

export const del = async (endpoint, token) => {
    try {
        const response = await fetch(
            `${apiConfig.protocol}://${apiConfig.address}/api/v${apiConfig.version}${endpoint}`,
            {
                method: "DELETE",
                headers: {
                    "Authorization": `Bearer ${token}`
                }
            }
        );
        if (!response.ok) {
            throw response.status;
        }
        return 200;
    } catch (e) {
        console.error(`Request error, statuscode: ${e}`);
        return {error: e};
    }
}