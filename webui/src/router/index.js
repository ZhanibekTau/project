import {createRouter, createWebHashHistory} from 'vue-router'
import ConversationView from "../views/ConversationView.vue";

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/user/:id',
			component: ConversationView,
			name: 'Conversation',
		},
	]
})

export default router
