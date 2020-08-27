"use strict";
//
import "./x/index.js";
//
import { setDataBinding } from "./util.js";
import * as ui from "./ui.js";
import { el_2, el_3, el_1, output, el_uonline, el_input, msg_processors } from "./ui.util.js";
import * as api from "./api/index.js";
import * as ws from "./ws.js";

//
$(document).on("click", (e) => {
    const p = e.target.path();
    if (p.filter((v) => v.matches("dialog[open]")).length > 0) return;
    if (p.filter((v) => v.matches(".msg .usr")).length > 0) return;
    const s = document.querySelectorAll("dialog[open]");
    s.forEach((v) => v.removeAttribute("open"));
});
window.addEventListener("blur", () => {
    ui.volatile.windowActive = false;
});
window.addEventListener("focus", () => {
    ui.volatile.windowActive = true;
});
document.getElementById("shrink_uonline").addEventListener("click", () => {
    output.classList.toggle("extended-right");
    el_uonline.toggleAttribute("hidden");
    el_input.classList.toggle("extended-right");
    output.children[0]._scroll_to_bottom();
});

//
(async function() {
    //
    moment.defaultFormat = "ddd MMM DD Y HH:mm:ss zZZ";

    //
    await api.M.meta.about().then((x) => {
        setDataBinding("server_name", x.name);
        setDataBinding("server_version", x.version);
        //
        const sx = document.querySelector("x-settings[data-s-for=server] [data-s-section=overview]");
        const sxc = sx.querySelectorAll("[endpoint][name]");
        for (const item of sxc) {
            const n = item.getAttribute("name");
            sx.querySelector(`[name="${n}"]`).setAttribute("value", x[n]);
        }
        //
        el_2.children[1].addEventListener("click", () => {
            document.querySelector("x-settings[data-s-for=server]")._open();
        });
    });

    //
    await api.M.users.me().then((x) => {
        ui.volatile.me = x.user;
        ui.volatile.me.perms = x.perms;
        const n = ui.volatile.me.nickname || ui.volatile.me.name;
        el_3.children[0].textContent = `@${n}`;
        document.querySelectorAll("[data-requires]").forEach((el) => {
            el.setAttribute("hidden", "");
        });
        const p = x.perms;
        for (const key in p) {
            if (p[key]) {
                document.querySelectorAll(`[data-requires^="${key}"]`).forEach((el) => {
                    el.removeAttribute("hidden");
                });
            }
        }
        //
        el_3.children[1].addEventListener("click", () => {
            document.querySelector("x-settings[data-s-for=user]")._open();
        });
        //
        document.querySelectorAll("[data-s-for='user'] [data-s-section='my_account'] [fill]").forEach((el) => {
            el.setAttribute("fill", x.user.uuid);
        });
        document.querySelector("[data-s-for='user'] [data-s-section='my_account'] [name='nickname']").setAttribute("value", x.user.nickname);
        //
        if (Notification.permission !== "granted") {
            ui.toggleHandlers.get("notifications_messages")("0");
        }
    }).catch(() => {
        location.assign("../");
    });

    //
    await api.M.roles.me().then((x) => {
        const rls = x.sort((a,b) => a.position > b.position);
        //
        for (const item of rls) {
            ui.M.role.add(item);
        }
    });

    //
    await api.M.channels.me().then(async (x) => {
        for (const item of x) {
            await ui.addChannel(item);
        }
        await output.setActiveChannel(x[0].uuid);

        el_1.querySelector("button").addEventListener("click", async () => {
            const {value: name} = await Swal({
                title: "Enter the new channel's name",
                input: "text",
                showCancelButton: true,
                inputValidator: (value) => !value && "You need to write something!",
            });
            if (name !== undefined) {
                return api.M.channels.create(name);
            }
        });
        el_1.querySelector("ol").addEventListener("click", (ev) => {
            const fl = ev.composedPath().filter((v) => v instanceof Element && v.matches("[data-uuid]"));
            if (fl.length === 0) return;
            output.setActiveChannel(fl[0].dataset.uuid);
        });
    });

    await api.M.invites.me().then((x) => {
        for (const item of x) {
            ui.M.invite.add(item);
        }
    });

    await api.M.users.online().then((x) => {
        for (const item of x) {
            ui.M.user.connect(item);
        }
    });

    //
    /** @type {HTMLInputElement} */
    const input = document.getElementById("input").children[0];
    const socket = new WebSocket(`ws${location.protocol.substring(4)}//${location.host}/ws`);

    socket.addEventListener("open", () => {
        el_2.children[0].classList.remove("loading");
        el_2.children[0].classList.add("online");
    });
    socket.addEventListener("close", () => {
        el_2.children[0].classList.remove("online");
    });
    socket.addEventListener("message", async (e) => {
        const d = JSON.parse(e.data);
        let o = ws.M;
        for (const item of d.type.split("-")) {
            if (!(item in o)) {
                console.error("event handler not found:", d);
                return;
            }
            o = o[item];
        }
        if (typeof o !== "function") {
            console.error("handler is not a function:", `"${d.type}"`, `"${typeof o}"`);
            return;
        }
        await o(d);
    });
    setInterval(() => {
        if (el_2.children[0].classList.contains("online")) {
            socket.send(JSON.stringify({
                type: "ping",
            }));
        }
    }, 30*1000);

    let processor_options = [];

    input.addEventListener('keydown', function(e) {
        const i = processor_options.findIndex(element => element[2] == true);
        if (processor_options.length != 0) {
            switch (event.key) {
                case "ArrowUp":
                    if (i > 0) {
                        processor_options[i][2] = false;
                        processor_options[i - 1][2] = true;
                    } else {
                        processor_options[i][2] = false;
                        processor_options[processor_options.length - 1][2] = true;
                    }
                    break;
                case "ArrowDown":
                    if (i < processor_options.length - 1) {
                        processor_options[i][2] = false;
                        processor_options[i + 1][2] = true;
                    } else {
                        processor_options[i][2] = false;
                        processor_options[0][2] = true;
                    }
                    break;
            }
            draw_suggestions(processor_options);
        }
    });

    input.addEventListener('input', (e) => {
        let input = e.target.value.toLowerCase();
        if (input[0] === '/' || input[0] === ':'){
            // get possible matches
            processor_options = [];
            for (const processor of msg_processors) {
                if (processor[0].toLowerCase().indexOf(input) > -1) {
                    processor.push(false);
                    processor_options.push(processor);
                }
            }
            // shorten list and set first option to as suggested
            if (processor_options.length > 10)
                processor_options = processor_options.slice(0, 10);
            if (processor_options.length > 0)
                processor_options[0][2] = true;
        } else {
            // if no command remove suggestions
            processor_options = []
        }
        draw_suggestions(processor_options)

    });

    //
    input.addEventListener("keydown", (e) => {
        if (e.key === "Enter") {
            let msg_con = "";
            if (processor_options.length) {
                msg_con = processor_options.find(element => element[2] == true)[0];
                processor_options = [];
            } else {
                msg_con = e.target.value;
            }
            if (msg_con.length === 0) return;
            socket.send(JSON.stringify({
                type: "message",
                in: output.active_channel_uid,
                message: msg_con.trim(),
            }));
            e.target.value = "";
            draw_suggestions(processor_options)
        }
    });
})();

function draw_suggestions(processor_options) {
    // check if suggestion box open
    let suggestions = document.getElementById("suggestions");
    let ul;
    if (!suggestions) {
        suggestions = document.createElement("div");
        suggestions.setAttribute("id", "suggestions");
        ul = document.createElement("ul");
    } else {
        ul = suggestions.children[0];
    }
    while(ul.firstChild) ul.removeChild(ul.firstChild);

    // create suggestion box
    for (const processor of processor_options) {
        let li = document.createElement("li");
        li.appendChild(document.createTextNode(processor[0]));
        let b = document.createElement("b");
        b.appendChild(document.createTextNode(processor[1]))
        if (processor[2])
            li.setAttribute("id", "selected-suggestion");
        li.appendChild(b);
        ul.appendChild(li)
    }

    suggestions.appendChild(ul);
    document.body.appendChild(suggestions);
}