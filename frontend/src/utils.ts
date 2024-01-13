import type { Ref } from "vue";

export async function makeGetRequest(url: string, busyRef: Ref<boolean>): Promise<any> {
    try {
        busyRef.value = true;
        const response = await fetch(url, { credentials: "include" });
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
        const response = await fetch(url, {
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
