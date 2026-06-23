"use client";

import Axios from "axios";

export default function Login() {

    const handleLogin = () => {
        Axios.get("http://localhost:8080/api/login-url").then((res) => {
            const loginURL = res.data.url;

            window.location.href = res.data.url;
        })
    }

    return (
        <div>
            <button onClick={() => {handleLogin()}}>Login</button>
        </div>
    )
}