const form = document.getElementById("shortenForm");
const urlInput = document.getElementById("url");
const aliasInput = document.getElementById("alias");

const result = document.getElementById("result");
const shortUrl = document.getElementById("shortUrl");
const openBtn = document.getElementById("openBtn");
const copyBtn = document.getElementById("copyBtn");
const statsBtn = document.getElementById("statsBtn");
const clicks = document.getElementById("clicks");
const message = document.getElementById("message");

let currentCode = "";

form.addEventListener("submit", async function (e) {
    e.preventDefault();

    message.textContent = "";
    result.classList.add("hidden");

    try {
        const response = await fetch("/shorten", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                url: urlInput.value,
                alias: aliasInput.value
            })
        });

        const data = await response.json();

        if (!response.ok) {
            message.textContent = data.error || "Something went wrong.";
            return;
        }

        currentCode = data.code;
        shortUrl.value = data.short_url;
        openBtn.href = data.short_url;
        clicks.textContent = "0";

        result.classList.remove("hidden");

    } catch (error) {
        message.textContent = "Could not connect to the server.";
    }
});

copyBtn.addEventListener("click", async function () {
    try {
        await navigator.clipboard.writeText(shortUrl.value);
        copyBtn.textContent = "Copied!";

        setTimeout(function () {
            copyBtn.textContent = "Copy";
        }, 1500);
    } catch (error) {
        shortUrl.select();
        document.execCommand("copy");
    }
});

statsBtn.addEventListener("click", async function () {
    if (!currentCode) {
        return;
    }

    try {
        const response = await fetch("/api/stats/" + currentCode);
        const data = await response.json();

        if (response.ok) {
            clicks.textContent = data.clicks;
        }
    } catch (error) {
        message.textContent = "Could not load statistics.";
    }
});
