document
    .getElementById("btn-signup")
    .addEventListener("click", async function(e) {
        e.preventDefault();
        await fetch(someFunction(window.location.href), {
                method: "POST",
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    username: document.getElementById("username").value.trim(),
                    email: document.getElementById("email").value.trim(),
                    password: document.getElementById("password").value.trim(),
                }),
            })
            .then((res) => {
                if (res.ok) {
                    return res.json();
                } else return Promise.reject(response);
            })
            .then((res) => console.log(res))
            .catch((err) => {
                console.log(err);
            });
    });

function someFunction(site) {
    return site.replace(/\/$/, "");
}