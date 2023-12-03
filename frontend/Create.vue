<script setup>
import { ref } from "vue";

const props = defineProps({ modelValue: Boolean });
const emit = defineEmits(["update:modelValue"]);

let createDialogURL = ref("");
let newShortLinkURL = ref("");

async function createLink() {
    const response = await fetch("/private/api/links", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ url: createDialogURL.value }),
    });
    const data = await response.json();
    newShortLinkURL.value = data.short_url;
    createDialogURL.value = "";
}

function closeDialog() {
    emit("update:modelValue", false);
    createDialogURL.value = "";
    newShortLinkURL.value = "";
}
</script>

<template>
    <dialog :open="props.modelValue">
        <article>
            <header>
                <a href="#" class="close" @click.prevent="closeDialog"></a>
                <div>New Short Link</div>
            </header>
            <form @submit.prevent="createLink()">
                <input
                    type="url"
                    v-model="createDialogURL"
                    placeholder="Enter a link (e.g. https://example.com/...)"
                    required
                />
                <button type="submit">Shorten Link</button>
            </form>
            <footer :hidden="!newShortLinkURL">
                <a :href="newShortLinkURL" target="_blank">{{ newShortLinkURL }}</a>
                <ins hidden>Copied</ins>
            </footer>
        </article>
    </dialog>
</template>

<style scoped>
dialog > article {
    width: 50%;
}
dialog > article > footer {
    text-align: left;
}
</style>
