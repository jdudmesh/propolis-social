/** @type {import('./$types').Actions} */
export const actions = {
	post: async ({ cookies, request }) => {
    const data = await request.formData();
    console.log(data);
    return { success: true };
  },
};