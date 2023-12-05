<script setup>
import { ref, watch } from "vue";
import { makePostRequest } from "./utils.js";
import Clipboard from "./Copier.vue";

const emit = defineEmits(["closed"]);

let open = ref(false);
let busy = ref(false);
let createDialogURL = ref("");
let newShortLinkURL = ref("");

async function createLink() {
    const data = await makePostRequest("/private/api/links", { url: createDialogURL.value }, busy);
    newShortLinkURL.value = data["short_url"];
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
    <button type="button" class="open" @click="open = true" v-bind="$attrs">Create New Short Link</button>
    <dialog :open="open">
        <article>
            <header>
                <a href="#" class="close" @click.prevent="open = false"></a>
                <div>Create New Short Link</div>
            </header>
            <form @submit.prevent="createLink()">
                <input
                    type="url"
                    v-model="createDialogURL"
                    placeholder="Enter a link (e.g. https://example.com/...)"
                    required
                />
                <button type="submit" :aria-busy="busy">Shorten Link</button>
            </form>
            <footer :hidden="!newShortLinkURL">
                <a :href="newShortLinkURL" target="_blank">{{ newShortLinkURL }}</a>
                <Clipboard :text="newShortLinkURL" :key="newShortLinkURL" />
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
footer > a {
    margin-right: 0.5rem;
}
</style>
