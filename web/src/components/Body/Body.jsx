import HabitsList from "../HabitsList/HabitsList";
import Auth from "../Auth/Auth";
import { clearToken, getToken } from "../storage_utils";
import "./Body.css";
import { useState } from "react";

export default function Body(props) {
    const [isLoggedIn, setLogged] = useState(getToken() !== null);

    const onLoginSuccess = () => {
        setLogged(true);
    }

    const onLogout = () => {
        clearToken();
        setLogged(false);
    }

    return (
        <main className="main">
            { isLoggedIn ?
                <> 
                    <h2 className="habits__title">Ваш список привычек</h2>
                    <HabitsList onAuthFailed={onLogout}></HabitsList>
                    <button className="main__logout-button" onClick={onLogout}>Выйти</button>
                </>
                : 
                <>
                    <Auth onLoginSuccess={onLoginSuccess}></Auth>
                </>
            }
        </main>
    );
}

