# python3 fix_duplicate_members.py --host XXXXX --user XXXXX --password XXXXX --dry_run
# dyr_run will make sure data is not deleted, but will show all the data that will get deleted
# this script also generates an SQL file with insert queries for each of the records which will be deleted
# We do not deleted any records having teams
# Retain the latest entry

import argparse
import logging
import os
import time

import pymysql


class MySqlConnector:
    def __init__(self, db_name, db_host, db_user, db_password):
        self.db_connection = None
        self.db_name = None
        self.get_mysql_db_connection(db_name, db_host, db_user, db_password)

    def get_mysql_db_connection(self, db_name, db_host, db_user, db_password):
        LOGGER.info("Mysql open connection")
        con = pymysql.connect(
            host="dbprodclone.signeasy.io",
            user="signeasy",
            password="se10%27$2012sy",
            database="se_purchases",
            cursorclass=pymysql.cursors.DictCursor,
        )
        self.db_connection = con

    def fetch_one(self, query, args):
        try:
            with self.db_connection.cursor() as cursor:
                cursor.execute(query, args)
                return cursor.fetchone()
        except Exception as e:
            LOGGER.error("Could not check if user is valid")
            LOGGER.exception(e)
            return False

    def fetch_all(self, query, args=()):
        try:
            with self.db_connection.cursor() as cursor:
                cursor.execute(query, args)
                return cursor.fetchall()
        except Exception as e:
            LOGGER.error("Could not check if user is valid")
            LOGGER.exception(e)
            return False

    def execute(self, query_list, commit=False):
        try:
            with self.db_connection.cursor() as cursor:
                for query in query_list:
                    # LOGGER.info(query)
                    cursor.execute(query)
                if commit:
                    self.db_connection.commit()
        except Exception as e:
            LOGGER.error(f"Error while executing query : {query}")
            self.db_connection.rollback()
            raise e

    def __del__(self):
        if self.db_connection and self.db_connection.open:
            self.db_connection.close()


def delete_records(mysql_con, account_id, member_id, member_association_id, is_dry_run):
    account_delete_query = f"DELETE FROM account WHERE id={account_id}"
    member_delete_query = f"DELETE FROM member WHERE id={member_id}"
    member_association_delete_query = f"DELETE FROM member_association WHERE id={member_association_id}"

    mysql_con.execute(
        [member_association_delete_query, member_delete_query, account_delete_query], commit=not is_dry_run
    )


def delete_duplicates(db_host, db_user, db_password, dry_run):
    mysql_con = MySqlConnector('se_user', db_host, db_user, db_password)

    duplicate_user_query = "select user_id, status from se_user.member where status='ACCPTD' group by user_id HAVING count(*) >1"

    duplicate_user_dict_list = mysql_con.fetch_all(duplicate_user_query)
    LOGGER.info(f"Total duplicate accounts found: {len(duplicate_user_dict_list)}")

    file_new = open(f"deleted_records_{int(time.time())}.sql", "w")

    for user_info in duplicate_user_dict_list:
        LOGGER.info(f"Finding duplicates for user: {user_info['user_id']}")
        query = f"SELECT mas.id association_id, mem.id member_id, mem.account_id, mas.team_id FROM member mem INNER JOIN member_association mas ON mem.id=mas.member_id WHERE user_id={user_info['user_id']} AND status='ACCPTD'  order by account_id DESC"
        user_duplicate_list = mysql_con.fetch_all(query)

        skip_list = []
        delete_list = []
        for dup_user in user_duplicate_list[1:]:
            if dup_user["team_id"]:
                skip_list.append(dup_user)
            else:
                delete_list.append(dup_user)

        if skip_list:
            if len(skip_list) > 1:
                LOGGER.error(
                    f"Need to manually delete the duplicates for this user: {user_info['user_id']}. These records have teams: {skip_list[0]}")
                continue

            if user_duplicate_list[0]["team_id"]:
                LOGGER.error(
                    f"Need to manually delete the duplicates for this user: {user_info['user_id']}, These records have teams: {user_duplicate_list[0]},{skip_list[0]}")
                continue
            delete_list.append(user_duplicate_list[0])


        if skip_list:
            LOGGER.info(f"Retaining these records for user account_id: {skip_list[0]['account_id']}, member_id: {skip_list[0]['member_id']}, member_association_id: {skip_list[0]['association_id']}")
        else:
            LOGGER.info(f"Retaining these records for user account_id: {user_duplicate_list[0]['account_id']}, member_id: {user_duplicate_list[0]['member_id']}, member_association_id: {user_duplicate_list[0]['association_id']}")

        LOGGER.info(f"Deleting these records for the user: {user_info['user_id']}")

        for delete_user in delete_list:
            account_insert_query = "INSERT into account (id, name, status, created_at, modified_at) VALUES(%s, '%s', '%s', '%s', '%s');\n"
            member_insert_query = "INSERT INTO member (id, account_id, user_id, status, invite_code, created_at, modified_at) VALUES(%s, %s, %s, '%s', '%s', '%s', '%s');\n"
            mewber_asoc_insert_query = "INSERT INTO member_association (id, member_id, role_id, created_at, modified_at) VALUES(%s, %s, %s, '%s', '%s');\n"
            LOGGER.info(
                f"DELETING account_id: {delete_user['account_id']}, member_id: {delete_user['member_id']}, member_association_id: {delete_user['association_id']}")
            account_query = f"SELECT * from account where id={delete_user['account_id']}"
            member_query = f"SELECT * from member where id={delete_user['member_id']}"
            member_association_query = f"SELECT * from member_association where id={delete_user['association_id']}"

            result = mysql_con.fetch_one(account_query, ())
            query = account_insert_query % (
            result["id"], result["name"], result["status"], result["created_at"], result["modified_at"])
            file_new.write(query)
            result = mysql_con.fetch_one(member_query, ())
            query = member_insert_query % (result["id"], result["account_id"], result["user_id"], result["status"], result["invite_code"], result["created_at"], result["modified_at"])
            file_new.write(query)
            result = mysql_con.fetch_one(member_association_query, ())
            query = mewber_asoc_insert_query % (result["id"], result["member_id"], result["role_id"], result["created_at"], result["modified_at"])
            file_new.write(query)
            delete_records(mysql_con, delete_user['account_id'], delete_user['member_id'], delete_user['association_id'], dry_run)
            file_new.flush()
    file_new.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Script to update is_admin column of members table')
    parser.add_argument('--log_level', default="INFO", choices=["INFO", "DEBUG", "WARNING", "ERROR"])
    parser.add_argument("--host", help="database host", default=os.environ.get("DB_HOST"), type=str)
    parser.add_argument("--user", help="database user", default=os.environ.get("DB_USER"), type=str)
    parser.add_argument("--password", help="database password", default=os.environ.get("DB_PASSWORD"), type=str)
    parser.add_argument("--dry_run", help="chargbee csv file", action="store_true")

    args = parser.parse_args()

    host = args.host
    user = args.user
    password = args.password
    dry_run = args.dry_run

    log_format = "[%(asctime)s] %(levelname)s  %(filename)s  Line: %(lineno)d  %(message)s"

    logging.basicConfig(filename=f"./log_delete.log",
                        filemode='a',
                        format=log_format,
                        datefmt='%H:%M:%S',
                        level=args.log_level)

    LOGGER = logging.getLogger(__name__)
    consoleHandler = logging.StreamHandler()
    logFormatter = logging.Formatter(log_format)
    consoleHandler.setFormatter(logFormatter)
    LOGGER.addHandler(consoleHandler)

    delete_duplicates(host, user, password, dry_run)
