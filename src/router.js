import { createWebHistory, createRouter } from "vue-router";

const routes = [
    // {
    //     path: "/",
    //     alias: "/home",
    //     name: "Главная",
    //     component: () => import("./components/HomePage")
    // },
    {
        path: "/",
        name: "Льготники",
        component: () => import("./components/RetireesList")
    },
    {
        path: "/updates",
        name: "Обновления",
        component: () => import("./components/UpdatesList")
    },
    {
        path: "/breakers",
        name: "Нарушители",
        component: () => import("./components/BreakersList")
    },
    // {
    //     path: "/mails",
    //     name: "Почтовые логи",
    //     component: () => import("./components/MailsList")
    // }
]

const router = createRouter({
    history: createWebHistory(),
    routes
});

export default router;