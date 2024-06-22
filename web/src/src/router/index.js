import { createWebHistory, createRouter } from 'vue-router'
import index from '~/view/index.vue'
import setting from '~/view/setting.vue'

const routes = [
    {
        name: "index",
        path: "/",
        components: {
            default: index
        }
    },
    {
        name: "setting",
        path: "/setting",
        components: {
            default: setting
        }
    }
]

export default createRouter({
    history: createWebHistory(),
    routes,
})