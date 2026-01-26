import HabitsList from "../HabitsList/HabitsList";
import Auth from "../Auth/Auth";
import { getToken } from "../storage_utils";
import "./Body.css";

export default function Body(props) {
    return (
        <main className="main">
            { getToken() !== null ?
                <> 
                    <h2 className="habits__title">Ваш список привычек</h2>
                    <HabitsList></HabitsList>
                </>
                : 
                <>
                    <Auth></Auth>
                </>
            }
        </main>
    );
}

