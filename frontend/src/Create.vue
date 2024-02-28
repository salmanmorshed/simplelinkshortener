<script setup lang="ts">
import { ref, watch } from "vue";
import type { NewLink } from "./types";
import { makePostRequest } from "./utils";
import Clipboard from "./Copier.vue";

const emit = defineEmits<{ (e: "closed"): void }>();

let open = ref(false);
let busy = ref(false);
let createDialogURL = ref("");
let newShortLinkURL = ref("");

async function createLink() {
    const data = (await makePostRequest("/api/links", { url: createDialogURL.value }, busy)) as NewLink;
    newShortLinkURL.value = data.short_url;
    createDialogURL.value = "";
}

watch(open, val => {
    if (!val) {
        createDialogURL.value = "";
        newShortLinkURL.value = "";
        emit("closed");
    }
});
</script>

<template>
    <button type="button" class="open" @click="open = true" v-bind="$attrs">Create new short link</button>
    <dialog :open="open">
        <article>
            <header>
                <button aria-label="Close" rel="prev" @click.prevent="open = false"></button>
                <p>
                    <strong>Create a new short link</strong>
                </p>
            </header>
            <form @submit.prevent="createLink()">
                <input
                    type="url"
                    v-model="createDialogURL"
                    placeholder="Enter a link (e.g. https://example.com/...)"
                    required
                />
                <button type="submit" :aria-busy="busy">Shorten link</button>
            </form>
            <footer :hidden="!newShortLinkURL">
                <a :href="newShortLinkURL" target="_blank">{{ newShortLinkURL }}</a>
                <Clipboard :text="newShortLinkURL" :key="newShortLinkURL" />
            </footer>
        </article>
    </dialog>
</template>

<style scoped>
dialog > article > footer {
    text-align: left;
}
footer > a {
    margin-right: 0.5rem;
}
</style>