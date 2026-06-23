"use client"

import React from "react";
import { useState, useEffect, useRef } from "react";

export default function Dashboard() {

    const [price, setPrice] = useState(0);
    const [status, setStatus] = useState("Connecting to Live Feed...");
    const wsRef = useRef(null);

    useEffect(() => {
        wsRef.current = new WebSocket("ws://localhost:8080/ws");

        wsRef.current.onopen = () => {
            setStatus("🟢 Live Market Data Active");
        };

        wsRef.current.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.last_price) {
                setPrice(data.last_price);
            }
        };

        wsRef.current.onclose = () => {
            setStatus("🔴 Disconnected from Server");
        };

        return () => {
            if (wsRef.current) wsRef.current.close();
        };
    }, [])

    return (
        
        <div>
            <h1>Dashboard</h1>
            <p>{status}</p>
            <p>Last Price: {price}</p>
        </div>
    )
}