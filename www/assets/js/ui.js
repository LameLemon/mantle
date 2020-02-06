/**
 */
//
import { el_1, el_2, el_3, create_element, dcTN, messageCache, output, getUserFromUUID, el_4 } from "./util.js";

//

export const volatile = {
    activeChannel: null,
};

//

export function addChannel(uuid, name) {
    el_1.firstElementChild.appendChild(create_element("li", [["data-uuid",uuid]], [
        create_element("div", [], [dcTN(name)]),
    ]))
    messageCache.set(uuid, []);
}

export async function addMessage(channel=volatile.activeChannel.dataset.uuid, from, message, raw_from=false, save=true, at=Date.now()) {
    const at_bottom = output.scrollTop === output.scrollTopMax;
    const uuid = raw_from ? "" : from.uuid;
    const name = from.nickname || from.name;
    const time = new Date(at).toLocaleString();
    if (raw_from || output.dataset.active === channel) {
        output.appendChild(create_element("div", [["class","msg"]], [
            create_element("div", [["class","ts"],["title",time]], [dcTN(time.substring(time.indexOf(" ")))]),
            create_element("div", [["class","usr"]], [dcTN(name + ": ")]),
            create_element("div", [["class","dat"]], [dcTN(message)])
        ]));
    }
    if (at_bottom) output.scrollTop = output.scrollHeight;
    if (save===true) messageCache.get(channel).push([uuid, message]);
}

export function getChannel(uid) {
    return el_1.querySelector(`[data-uuid="${uid}"]`);
}

export async function setActiveChannel(uid) {
    console.debug("channel-switch:", uid);
    let ac = el_1.querySelector(".active");
    if (ac !== null) ac.classList.remove("active");
    const c = getChannel(uid);
    volatile.activeChannel = c;
    volatile.activeChannel.classList.add("active");
    //
    output.dataset.active = uid;
    output.removeAllChildren();
    const new_message_history = messageCache.get(uid);
    for (const item of new_message_history) {
        addMessage(uid, await getUserFromUUID(item[0]), item[1], false, false);
    }
}

export async function setMemberOnline(uid) {
    console.debug("user-ws-connect", uid);
    const ue = el_4.querySelector(`[data-user="${uid}"]`);
    if (ue === null) {
        const u = await getUserFromUUID(uid);
        el_4.appendChild(create_element("li", [["data-user",uid]], [
            dcTN(u.name)
        ]));
    }
}

export async function setMemberOffline(uid) {
    console.debug("user-ws-disconnect", uid);
    const ue = el_4.querySelector(`[data-user="${uid}"]`);
    if (ue !== null) {
        ue.remove();
    }
}
