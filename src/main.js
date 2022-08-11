import { createApp } from 'vue'
import App from './App.vue'
import router from "./router"
import ElementPlus from 'element-plus';
import 'element-plus/lib/theme-chalk/index.css';
import moment from 'moment'
import './assets/main.css';
import locale from 'element-plus/lib/locale/lang/ru'
import 'dayjs/locale/ru'

// Load Locales ('en' comes loaded by default)
require('moment/locale/ru');
// Choose Locale
moment.locale('ru');
// Vue.prototype.moment = moment

createApp(App).use(router).use(ElementPlus, { locale }).use(moment).mount('#app')
