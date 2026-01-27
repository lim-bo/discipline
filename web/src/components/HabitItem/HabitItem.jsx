import "./HabitItem.css";

export default function HabitItem(props) {
    return (
        <li className="habits__item">
            <h3 className="habits__item-title">{props.title}</h3>
            <p className="habits__item-description">{
                props.description ? props.description : "No description..."
            }</p>
            <form className="habits__item-form">
                <input className="habits__item-check" type="checkbox"/>
                <button className="habits__item-delete-button" onClick={props.onDelete}>Удалить</button>
            </form>
        </li>
    );
}