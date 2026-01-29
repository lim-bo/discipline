import "./HabitItem.css";

export default function HabitItem(props) {

    const handleDeletion = async (event) => {
        event.preventDefault();
        props.onDelete(props.id);
    }

    return (
        <li className="habits__item">
            <h4 className="habits__item-title">{props.title}</h4>
            <p className="habits__item-description">{
                props.description ? props.description : ""
            }</p>
            <form className="habits__item-form">
                <label className="habits__item-check-wrap button-hover-accent">
                    <input id="habit-check" className="habits__item-check" type="checkbox"/>
                    <svg xmlns="http://www.w3.org/2000/svg" className="check-icon" viewBox="0 0 24 24">
                        <path d="M22.319 4.431L8.5 18.249a1 1 0 01-1.417 0L1.739 12.9a1 1 0 00-1.417 0 1 1 0 000 1.417l5.346 5.345a3.008 3.008 0 004.25 0L23.736 5.847a1 1 0 000-1.416 1 1 0 00-1.417 0z"/>
                    </svg>
                </label>
                <button className="habits__item-delete-button button-hover-accent" onClick={handleDeletion}>
                    <svg xmlns="http://www.w3.org/2000/svg" className="habits__item-delete-icon" viewBox="0 0 24 24">
                        <path d="M21 4h-3.1A5.009 5.009 0 0013 0h-2a5.009 5.009 0 00-4.9 4H3a1 1 0 000 2h1v13a5.006 5.006 0 005 5h6a5.006 5.006 0 005-5V6h1a1 1 0 000-2zM11 2h2a3.006 3.006 0 012.829 2H8.171A3.006 3.006 0 0111 2zm7 17a3 3 0 01-3 3H9a3 3 0 01-3-3V6h12z"/>
                        <path d="M10 18a1 1 0 001-1v-6a1 1 0 00-2 0v6a1 1 0 001 1zM14 18a1 1 0 001-1v-6a1 1 0 00-2 0v6a1 1 0 001 1z"/>
                    </svg>
                </button>
            </form>
        </li>
    );
}