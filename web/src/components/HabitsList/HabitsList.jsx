import { useCallback, useEffect, useState } from "react";
import HabitItem from "../HabitItem/HabitItem";
import "./HabitsList.css";
import { apiConfig, del, get, post } from "../client";
import { getToken } from "../storage_utils";

export default function HabitsList(props) {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [currentDate, setCurrentDate] = useState(new Date());
    const [addHabitData, setAddHabitForm] = useState({
        title: "",
        desc: ""
    });

    const onChangeAddForm = (e) => {
        setAddHabitForm((prev) => (
            {
                ...prev,
                [e.target.id]: e.target.value
            }
        ));
    }

    const handleFetchError = (status) => {
        if (status === 401) {
            props.onAuthFailed();
        }
    }

    const onAddHabit = useCallback(async (event) => {
        event.preventDefault();
        if (!addHabitData.title.trim()) {
            alert("Missing habit title");
            return;
        }

        const res = await post("/habits", getToken(), {
            title: addHabitData.title,
            desc: addHabitData.desc
        });

        if (res.habit_id) {
            setItems((prev) => (
                [...prev, 
                    {id: res.habit_id, title: addHabitData.title, desc: addHabitData.desc}
                ]
            ));
            event.target.reset();
            setAddHabitForm({title:"", desc:""});
        } else {
            handleFetchError(res.error);
        }

    }, [addHabitData]);

    const deleteHabit = useCallback(async (habitID) => {
        const res = await del(`${apiConfig.endpoints.deleteHabit}/${habitID}`, getToken());
        if (!res.error) {
            setItems((prev) => (prev.filter(item => item.id !== habitID)));
        }
    }, []);

    

    useEffect(() => {
        const fetchHabits = async () => {
            setLoading(true);
            const habits = await get("/habits", getToken());
            if (habits.habits) {
                const newItems = [];
                habits.habits.map((item) => {
                    newItems.push({id: item.id, title: item.title, desc: item.desc});
                });
                setItems(newItems);
            } else {
                handleFetchError(habits.error);
            }
            setLoading(false);
        }
        fetchHabits();
    }, []);

    return (
        <div className="habits">
            <h3>Сегодня: {currentDate.toLocaleDateString("ru-RU")}</h3>
            <ul className="habits__list">
                {
                    !loading ? (items.map((item) => (
                        <HabitItem
                            key={item.id}
                            id={item.id}
                            title={item.title}
                            description={item.desc}
                            onDelete={deleteHabit}
                            onFetchError={handleFetchError}
                            date={toISODate(currentDate)}
                        />
                    ))) 
                    : <p className="habits__list-loading">Загрузка</p>
                }
            </ul>
            <form className="habits__add-habit-form" onSubmit={onAddHabit}>
                <label className="habits__input-label">
                    <input id="title" className="habits__text-field" type="text" placeholder="Название" onChange={onChangeAddForm}/>
                </label>
                <label className="habits__input-label">
                    <input id="desc" className="habits__text-field" type="text" placeholder="Описание" onChange={onChangeAddForm}/>
                </label>
                <button className="habits__add-button" type="submit">Добавить</button>
            </form>
        </div>
    );
}

const toISODate = (date) => {
    const offsetMinutes = date.getTimezoneOffset();
    const isoStringOffset = new Date(date.getTime() - offsetMinutes * 60000).toISOString();
    return isoStringOffset.slice(0, 10);
}