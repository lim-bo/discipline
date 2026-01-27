export const apiConfig = {
    protocol: "http",
    address: "localhost:8070",
    version: "1",
    endpoints: {
        login: "/auth/login",
        register: "/auth/register"
    }
}

export const post = async (endpoint, reqBody) => {
    try {
        const response = await fetch(
            `${apiConfig.protocol}://${apiConfig.address}/api/v${apiConfig.version}${endpoint}`,
            {
                method: "POST",
                body: JSON.stringify(reqBody)
            }
        );
        if (!response.ok) {
            throw new Error(response.status);
        }
        const body = await response.json();
        return body;
    } catch (e) {
        console.error(`Request error, statuscode: ${e}`);
        return { error: e };
    }    
}