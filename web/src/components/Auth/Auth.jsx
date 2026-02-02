import { useState } from "react";
import "./Auth.css";
import { apiConfig, post } from "../client";
import { setToken } from "../storage_utils";

export default function Auth({ onLoginSuccess }) {
    const [registrationChecked, changeMode] = useState(false);

    const [formData, setFormData] = useState(
        {
            username: "",
            password: ""
        }
    );

    const changeAuthMode = (event) => {
        changeMode(event.target.checked);
    }

    const handleChange = (event) => {
        setFormData((prev) => ({
            ...prev,
            [event.target.id]: event.target.value
        }));
    }

    const submitHandler = async (event) => {
        event.preventDefault();
        const requestEndpoint = registrationChecked ? apiConfig.endpoints.register : apiConfig.endpoints.login;
        const userPayload = await post(requestEndpoint, null, {
            name: formData.username,
            password: formData.password
        });
        if (userPayload.token) {
            setToken(userPayload.token);
            setFormData({
                username: "",
                password: ""
            });
            event.target.reset();
            onLoginSuccess();
        } else if (userPayload.uid) {
            alert("Registered")
            setFormData({
                username: "",
                password: ""
            });
            event.target.reset();
        } else {
            alert(`Register/Login error: ${userPayload.error}`);
        }        
    }

    return (
        <div className="auth">
            <form className="auth__form" onSubmit={submitHandler}>
                <label className="auth__input-label">
                    Имя пользователя
                    <input id="username" 
                        type="text" 
                        className="auth__input-text"
                        placeholder="HadrWorker2000"
                        onChange={handleChange}
                        required
                    />
                </label>
                <label className="auth__input-label">
                    Пароль
                    <input id="password" 
                        type="password" 
                        className="auth__input-text"
                        placeholder="secret_password"
                        onChange={handleChange}
                        required
                    />
                </label>
                <label className="auth__input-label">
                    Регистрация
                    <input id="auth-mode" 
                        type="checkbox" 
                        className="auth__input-checkbox" 
                        checked={registrationChecked} 
                        onChange={changeAuthMode}
                    />
                </label>
                <button type="submit" className="auth__submit">
                    { registrationChecked ? "Регистрация" : "Вход"}
                </button>
            </form>
        </div>
    );
}