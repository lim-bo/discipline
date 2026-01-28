import { useEffect, useState } from "react";
import HabitItem from "../HabitItem/HabitItem";
import "./HabitsList.css";

export default function HabitsList(props) {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);

    const refreshData = () => {
        fetchHabits();
    }

    const fetchHabits = async () => {
        try {
            setLoading(true);
            setItems(tempHabits);
        } catch {
        } finally {
            setLoading(false);
        }
    }

    const addHabit = async (newHabit) => {
        tempHabits.push({
            id: newHabit.id,
            title: newHabit.title,
            description: newHabit.description
        });
        fetchHabits();
    }

    const deleteHabit = async (e) => {
        e.preventDefault();
    }

    useEffect(() => {
        refreshData();
    }, [items]);

    return (
        <ul className="habits__list">
            {
                !loading ? (items.map((item) => (
                    <HabitItem
                        key={item.id}
                        title={item.title}
                        description={item.description}
                        onDelete={deleteHabit}
                    />
                ))) 
                : <p className="habits__list-loading">Загрузка</p>
            }
        </ul>
    );
}

const tempHabits = [
    {
        id: 1, 
        title: "Выпить 6 стаканов воды", 
        description: "Нужно выпить много воды для здорового обмена веществ", 
    },
    {
        id: 2, 
        title: "Утренняя молитва"
    },
    {
        id: 3, 
        title: "Принять витамины", 
        description: "Необходимо выпить омега-3 и цинк"
    }
]