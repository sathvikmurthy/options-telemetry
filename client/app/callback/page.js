"use client"

import React from "react";
import { useState, useEffect } from "react";
import Axios from "axios"; 
import { useRouter } from 'next/navigation'

export default function Callback() {
    const router = useRouter()

    const [status, setStatus] = useState("Waiting for token...");

    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get("request_token")

        if(token) {
            setStatus("Toke received, Auth with Backend...!");

            Axios.post("http://localhost:8080/api/start-session", {
                request_token: token
            }).then((res) => {
                setStatus("Successfully Logged in! Redirecting...");
                router.push("/dashboard")
            }).catch(err => {
                console.error(err);
                setStatus("Login failed. Check backend console.");
            });
        } else {
            setStatus("Waiting for authentication redirect...");
        }
    }, []);

    return (
        <div>
            <h2>{status}</h2>
        </div>
    )
}