import type { Ref } from "vue";

function debugURL(url: string): string {
    if (import.meta.env.PROD) return url;
    return (import.meta.env.VITE_API_HOST ?? "") + url;
}

export async function makeGetRequest(url: string, busyRef: Ref<boolean>): Promise<any> {
    try {
        busyRef.value = true;
        const response = await fetch(debugURL(url), { credentials: "include" });
        const data = await response.json();
        busyRef.value = false;
        return data;
    } catch (error) {
        busyRef.value = false;
        console.log(error);
    }
}

export async function makePostRequest(url: string, payload: any, busyRef: Ref<boolean>): Promise<any> {
    try {
        busyRef.value = true;
        const response = await fetch(debugURL(url), {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify(payload),
        });
        const data = await response.json();
        busyRef.value = false;
        return data;
    } catch (error) {
        busyRef.value = false;
        console.log(error);
    }
}

export function formatDateTime(input: string): string {
    return new Date(Date.parse(input)).toLocaleString();
}
