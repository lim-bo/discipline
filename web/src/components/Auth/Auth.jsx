import { useState } from "react";
import "./Auth.css";

export default function Auth() {
    const [registrationChecked, changeMode] = useState(false);

    const changeAuthMode = (event) => {
        changeMode(event.target.checked);
    }

    return (
        <div className="auth">
            <form className="auth__form">
                <label className="auth__input-label">
                    Имя пользователя
                    <input id="username" 
                        type="text" 
                        className="auth__input-text"
                        placeholder="HadrWorker2000"
                    />
                </label>
                <label className="auth__input-label">
                    Пароль
                    <input id="password" 
                        type="password" 
                        className="auth__input-text"
                        placeholder="secret_password"
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
                <button type="button" className="auth__submit">
                    { registrationChecked ? "Регистрация" : "Вход"}
                </button>
            </form>
        </div>
    );
}