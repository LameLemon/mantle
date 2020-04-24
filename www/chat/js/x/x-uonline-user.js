"use strict";
//
import { create_element, dcTN } from "./../util.js";
import * as api from "./../api/index.js";

//
customElements.define("x-uonline-user", class extends HTMLElement {
    constructor() {
        super();
    }
    async connectedCallback() {
        this._uid = this.getAttribute("uuid");
        const o = await api.M.users.get(this._uid);
        {
            const r = await o.getRoles();
            const l = r.filter((v) => v.color.length > 0);
            if (l.length > 0) {
                this.setAttribute("data-role", l[0].uuid);
            }
        }
        this.appendChild(create_element("span", null, [dcTN(o.name)]));
        this.appendChild(create_element("span", null, [dcTN("#"+o.id)]));
    }
    get role_element() {
        return this.parentElement.parentElement;
    }
    removeMe() {
        this.role_element.removeUser(this._uid);
    }
});
