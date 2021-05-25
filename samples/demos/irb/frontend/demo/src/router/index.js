import Vue from 'vue'
import VueRouter from 'vue-router'
import About from '../views/About.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'Home',
    component: About
  },
  {
    path: '/consenter',
    name: 'Consenter',
    component: () => import('../views/Consenter.vue')
  },
  {
    path: '/experimenter',
    name: 'Experimenter',
    component: () => import('../views/Experimenter.vue')
  },
  {
    path: '/approver',
    name: 'Approver',
    component: () => import(/* webpackChunkName: "about" */ '../views/Approver.vue')
  },

]

const router = new VueRouter({
  routes
})

export default router
