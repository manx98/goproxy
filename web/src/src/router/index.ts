import { createWebHistory, createRouter } from 'vue-router'

const routes = [
    {
        name: "index",
        path: "/",
        components: {
            default: import('~/view/index.vue')
        }
    },
]

export default createRouter({
    history: createWebHistory(),
    routes,
})