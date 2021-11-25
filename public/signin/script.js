document
    .getElementById("btn-signin")
    .addEventListener("click", async function(e) {
        e.preventDefault();
        await fetch(someFunction(window.location.href), {
                method: "POST",
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    email: document.getElementById("email").value.trim(),
                    password: document.getElementById("password").value.trim(),
                }),
            })
            .then((res) => {
                if (res.ok) {
                    return res.json();
                } else return Promise.reject(response);
            })
            .then((res) => {
                localStorage.setItem("username", res.username);
                localStorage.setItem("email", res.email);
                window.location.href = "/";
            })
            .catch((err) => {
                console.log(err);
            });
    });

function someFunction(site) {
    return site.replace(/\/$/, "");
}