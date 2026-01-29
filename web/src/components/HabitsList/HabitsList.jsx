import { useCallback, useEffect, useState } from "react";
import HabitItem from "../HabitItem/HabitItem";
import "./HabitsList.css";
import { get, post } from "../client";
import { getToken } from "../storage_utils";

export default function HabitsList(props) {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [currentDate, setCurrentDate] = useState(getCurrentDate);
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

    const onAddHabit = async (event) => {
        event.preventDefault();
        const res = await post("/habits", {
            title: addHabitData.title,
            desc: addHabitData.desc
        }, getToken());
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

    }

    const deleteHabit = async (habitID) => {
        e.preventDefault();
    }

    useEffect(() => {
        const fetchHabits = async () => {
            setLoading(true);
            const habits = await get("/habits", getToken());
            if (habits.habits) {
                const newItems = [];
                habits.habits.map((item) => {
                    newItems.push({id: item.id, title: item.title, desc: item.desc});
                })
                setItems(newItems);
            } else {
                handleFetchError(habits.error);
            }
            setLoading(false);
        };
        fetchHabits();
    }, []);

    return (
        <div className="habits">
            <h3>Сегодня: {currentDate}</h3>
            <ul className="habits__list">
                {
                    !loading ? (items.map((item) => (
                        <HabitItem
                            key={item.id}
                            title={item.title}
                            description={item.desc}
                            onDelete={deleteHabit}
                        />
                    ))) 
                    : <p className="habits__list-loading">Загрузка</p>
                }
            </ul>
            <form className="habits__add-habit-form" onChange={onChangeAddForm} onSubmit={onAddHabit}>
                <label className="habits__input-label">
                    <input id="title" className="habits__text-field" type="text" placeholder="Название"/>
                </label>
                <label className="habits__input-label">
                    <input id="desc" className="habits__text-field" type="text" placeholder="Описание"/>
                </label>
                <button className="habits__add-button" type="submit">Добавить</button>
            </form>
        </div>
    );
}

const getCurrentDate = () => {
    const now = new Date();
    return now.toLocaleDateString("ru-RU");
}